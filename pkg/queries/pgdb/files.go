package pgdb

import (
	"context"
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

// AddFiles add files to packages
func (q *Queries) AddFiles(ctx context.Context, files []pgdb.FileParams) ([]pgdb.File, error) {

	currentTime := time.Now()
	var values []interface{}
	var inserts []string

	for index, row := range files {
		inserts = append(inserts, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			index*13+1,
			index*13+2,
			index*13+3,
			index*13+4,
			index*13+5,
			index*13+6,
			index*13+7,
			index*13+8,
			index*13+9,
			index*13+10,
			index*13+11,
			index*13+12,
			index*13+13,
		))

		etag := fmt.Sprintf("{\"checksum\": \"%s\", \"chunkSize\": \"%s\", \"sha256\": \"%s\"}", row.CheckSum, "32", row.Sha256)

		values = append(values, row.PackageId, row.Name, row.FileType.String(), row.S3Bucket, row.S3Key,
			row.ObjectType.String(), row.Size, etag, row.UUID.String(), processingState.Unprocessed.String(),
			uploadState.Uploaded.String(), currentTime, currentTime)

	}

	sqlInsert := "INSERT INTO files(package_id, name, file_type, s3_bucket, s3_key, " +
		"object_type, size, checksum, uuid, processing_state, uploaded_state, created_at, updated_at) VALUES "

	returnRows := "id, package_id, name, file_type, s3_bucket, s3_key, " +
		"object_type, size, checksum, uuid, processing_state, uploaded_state, created_at, updated_at"

	sqlInsert = sqlInsert +
		strings.Join(inserts, ",") +
		fmt.Sprintf("ON CONFLICT (uuid) DO UPDATE SET updated_at=EXCLUDED.updated_at RETURNING %s;", returnRows)

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

// UpdateBucketForFile updates the storage bucket as part of upload process and sets Status
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
			nofFoundError := &pgdb.ErrFileNotFound{}
			log.Println(nofFoundError.Error())
			return nofFoundError
		}

		multipleRowError := &pgdb.ErrMultipleRowsAffected{}
		log.Println(multipleRowError.Error())
		return multipleRowError
	}

	return nil

}
