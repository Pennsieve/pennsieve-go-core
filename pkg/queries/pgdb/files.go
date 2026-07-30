package pgdb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/objectType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/processingState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/uploadState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

// filesPerRowColumns is the number of columns bound per row in the AddFiles
// INSERT. Kept as a constant so the placeholder arithmetic and the values slice
// can't drift apart silently.
const filesPerRowColumns = 13

// jsonQuotedString returns s as a JSON string literal, including the surrounding
// quotes and with any characters that need escaping escaped.
//
// The checksum column holds a JSON document that was previously assembled with
// fmt.Sprintf and raw %s interpolation, so a value containing a double quote or
// a backslash produced malformed JSON. In practice these fields hold S3 ETags
// and base64 SHA-256 digests, so this is defensive rather than a live bug — but
// it is the only unparameterised value in this statement and it costs nothing to
// escape properly.
//
// Note this escapes the individual values rather than marshalling the whole
// object, deliberately: it keeps the emitted JSON byte-identical to what this
// function has always produced (including the spaces after the colons) for every
// value that did not need escaping, so nothing downstream that string-matches on
// this column changes behaviour.
func jsonQuotedString(s string) string {
	encoded, err := json.Marshal(s)
	if err != nil {
		// json.Marshal only fails on unsupported types; a string always encodes.
		// Fall back to an empty JSON string rather than emitting invalid JSON.
		log.Warnf("unable to JSON-encode checksum component: %v", err)
		return `""`
	}
	return string(encoded)
}

// buildAddFilesQuery assembles the multi-row upsert and its bound values.
// Extracted from AddFiles so the generated SQL can be asserted on without a
// database. Callers must ensure files is non-empty; an empty slice produces an
// empty VALUES list, which is not valid Postgres.
func buildAddFilesQuery(files []pgdb.FileParams, currentTime time.Time) (string, []interface{}) {
	values := make([]interface{}, 0, len(files)*filesPerRowColumns)
	inserts := make([]string, 0, len(files))

	for index, row := range files {
		placeholders := make([]string, filesPerRowColumns)
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", index*filesPerRowColumns+i+1)
		}
		inserts = append(inserts, "("+strings.Join(placeholders, ",")+")")

		etag := fmt.Sprintf("{\"checksum\": %s, \"chunkSize\": \"%s\", \"sha256\": %s}",
			jsonQuotedString(row.CheckSum), "32", jsonQuotedString(row.Sha256))

		values = append(values, row.PackageId, row.Name, row.FileType.String(), row.S3Bucket, row.S3Key,
			row.ObjectType.String(), row.Size, etag, row.UUID.String(), processingState.Unprocessed.String(),
			uploadState.Uploaded.String(), currentTime, currentTime)
	}

	returnRows := "id, package_id, name, file_type, s3_bucket, s3_key, " +
		"object_type, size, checksum, uuid, processing_state, uploaded_state, created_at, updated_at"

	sqlInsert := "INSERT INTO files(package_id, name, file_type, s3_bucket, s3_key, " +
		"object_type, size, checksum, uuid, processing_state, uploaded_state, created_at, updated_at) VALUES " +
		strings.Join(inserts, ",") +
		// Leading space matters: without it the statement reads ")ON CONFLICT".
		// Postgres happens to tolerate that because ")" terminates the token, but
		// it is needlessly fragile and unreadable in logs.
		fmt.Sprintf(" ON CONFLICT (uuid) DO UPDATE SET updated_at = EXCLUDED.updated_at WHERE files.s3_bucket = EXCLUDED.s3_bucket AND files.s3_key = EXCLUDED.s3_key RETURNING %s;", returnRows)

	return sqlInsert, values
}

// AddFiles add files to packages
func (q *Queries) AddFiles(ctx context.Context, files []pgdb.FileParams) ([]pgdb.File, error) {

	// An empty batch is a no-op, not an error. Without this guard the builder
	// emits "INSERT INTO files(...) VALUES  ON CONFLICT ..." — an empty VALUES
	// list — and Postgres rejects it with
	//
	//     pq: syntax error at or near "ON"
	//
	// which points at the SQL rather than at the caller that passed nothing.
	// That cost real debugging time in upload-service-v2: a warm Lambda whose
	// (then process-global) dedup map had already recorded every uuid in a
	// redelivered SQS batch filtered that batch down to zero files, and the
	// resulting failure looked like a broken statement rather than an empty
	// input.
	if len(files) == 0 {
		return []pgdb.File{}, nil
	}

	sqlInsert, values := buildAddFilesQuery(files, time.Now())

	//prepare the statement
	stmt, err := q.db.PrepareContext(ctx, sqlInsert)
	if err != nil {
		log.Error("ERROR: ", err)
		return nil, err
	}

	//goland:noinspection ALL
	defer stmt.Close()

	// format all values at once
	var allInsertedFiles []pgdb.File
	rows, err := stmt.Query(values...)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			log.Println(pqErr)
		}
		return nil, err
	}

	for rows.Next() {
		var currentRecord pgdb.File

		var fType string
		var oType string
		var pState string
		var uState string

		err = rows.Scan(
			&currentRecord.Id,
			&currentRecord.PackageId,
			&currentRecord.Name,
			&fType,
			&currentRecord.S3Bucket,
			&currentRecord.S3Key,
			&oType,
			&currentRecord.Size,
			&currentRecord.CheckSum,
			&currentRecord.UUID,
			&pState,
			&uState,
			&currentRecord.CreatedAt,
			&currentRecord.UpdatedAt,
		)

		if err != nil {
			log.Println("ERROR: ", err)
			return nil, err
		}

		currentRecord.FileType = fileType.Dict[fType]
		currentRecord.ObjectType = objectType.Dict[oType]
		currentRecord.ProcessingState = processingState.Dict[pState]
		currentRecord.UploadedState = uploadState.Dict[uState]

		allInsertedFiles = append(allInsertedFiles, currentRecord)
	}

	return allInsertedFiles, nil
}

// IsFilePublished checks whether a file identified by its UUID has been published
// by looking for a non-null published_s3_version_id.
func (q *Queries) IsFilePublished(ctx context.Context, uploadId string, organizationId int64) (bool, error) {
	queryStr := fmt.Sprintf("SELECT published_s3_version_id IS NOT NULL FROM \"%d\".files WHERE UUID=$1;", organizationId)
	var published bool
	err := q.db.QueryRowContext(ctx, queryStr, uploadId).Scan(&published)
	if err != nil {
		return false, fmt.Errorf("error checking published status for file %s: %w", uploadId, err)
	}
	return published, nil
}

// UpdateBucketForUnpublishedFile updates the storage bucket for a file, but only if
// the file has not been published (published_s3_version_id IS NULL). This prevents
// the upload-move workflow from overwriting a publish-bucket location in the DB.
// Returns ErrFileNotFound if the file does not exist.
// Returns ErrFileAlreadyPublished if the file was published (0 rows affected due to the guard).
func (q *Queries) UpdateBucketForUnpublishedFile(ctx context.Context, uploadId string, bucket string, s3Key string, organizationId int64) error {
	queryStr := fmt.Sprintf("UPDATE \"%d\".files SET s3_bucket=$1, s3_key=$2 WHERE UUID=$3 AND published_s3_version_id IS NULL;", organizationId)
	result, err := q.db.ExecContext(ctx, queryStr, bucket, s3Key, uploadId)
	if err != nil {
		log.Printf("Error updating the bucket location: %v", err)
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affectedRows == 0 {
		// Distinguish between "file doesn't exist" and "file is published" by checking if the file exists.
		var exists bool
		existsQuery := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM \"%d\".files WHERE UUID=$1);", organizationId)
		if err := q.db.QueryRowContext(ctx, existsQuery, uploadId).Scan(&exists); err != nil {
			return fmt.Errorf("error checking file existence: %w", err)
		}
		if exists {
			return &pgdb.ErrFileAlreadyPublished{}
		}
		return &pgdb.ErrFileNotFound{}
	}
	if affectedRows > 1 {
		return &pgdb.ErrMultipleRowsAffected{}
	}
	return nil
}

// UpdateBucketForFile updates the storage bucket as part of upload process.
func (q *Queries) UpdateBucketForFile(ctx context.Context, uploadId string, bucket string, s3Key string, organizationId int64) error {

	queryStr := fmt.Sprintf("UPDATE \"%d\".files SET s3_bucket=$1, s3_key=$2 WHERE UUID=$3;", organizationId)
	result, err := q.db.ExecContext(ctx, queryStr, bucket, s3Key, uploadId)

	msg := ""
	if err != nil {
		msg = fmt.Sprintf("Error updating the bucket location: %v", err)
		log.Println(msg)
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affectedRows != 1 {
		if affectedRows == 0 {
			return &pgdb.ErrFileNotFound{}
		}

		multipleRowError := &pgdb.ErrMultipleRowsAffected{}
		log.Println(multipleRowError.Error())
		return multipleRowError
	}

	return nil

}
