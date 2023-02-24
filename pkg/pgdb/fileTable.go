package pgdb

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pennsieve/pennsieve-go-core/pkg/core"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/objectType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/processingState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/uploadState"
	"log"
	"strings"
	"time"
)

type File struct {
	Id              string                          `json:"id"`
	PackageId       int                             `json:"package_id"`
	Name            string                          `json:"name"`
	FileType        fileType.Type                   `json:"file_type"`
	S3Bucket        string                          `json:"s3_bucket"`
	S3Key           string                          `json:"s3_key"`
	ObjectType      objectType.ObjectType           `json:"object_type"`
	Size            int64                           `json:"size"`
	CheckSum        string                          `json:"checksum"`
	UUID            uuid.UUID                       `json:"uuid"`
	ProcessingState processingState.ProcessingState `json:"processing_state"`
	UploadedState   uploadState.UploadedState       `json:"uploaded_state"`
	CreatedAt       time.Time                       `json:"created_at"`
	UpdatedAt       time.Time                       `json:"updated_at"`
}

type FileParams struct {
	PackageId  int                   `json:"package_id"`
	Name       string                `json:"name"`
	FileType   fileType.Type         `json:"file_type"`
	S3Bucket   string                `json:"s3_bucket"`
	S3Key      string                `json:"s3_key"`
	ObjectType objectType.ObjectType `json:"object_type"`
	Size       int64                 `json:"size"`
	CheckSum   string                `json:"checksum"`
	Sha256     string                `json:"sha256"`
	UUID       uuid.UUID             `json:"uuid"`
}

type ErrFileNotFound struct{}

func (m *ErrFileNotFound) Error() string {
	return "File does not exist in postgres"
}

type ErrMultipleRowsAffected struct{}

func (m *ErrMultipleRowsAffected) Error() string {
	return "Multiple files in files table were updated"
}

func (p *File) Add(db core.PostgresAPI, files []FileParams) ([]File, error) {

	currentTime := time.Now()
	var vals []interface{}
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

		vals = append(vals, row.PackageId, row.Name, row.FileType.String(), row.S3Bucket, row.S3Key,
			row.ObjectType.String(), row.Size, etag, row.UUID.String(), processingState.Unprocessed.String(),
			uploadState.Uploaded.String(), currentTime, currentTime)

	}

	sqlInsert := "INSERT INTO files(package_id, name, file_type, s3_bucket, s3_key, " +
		"object_type, size, checksum, uuid, processing_state, uploaded_state, created_at, updated_at) VALUES "

	returnRows := "id, package_id, name, file_type, s3_bucket, s3_key, " +
		"object_type, size, checksum, uuid, processing_state, uploaded_state, created_at, updated_at"

	sqlInsert = sqlInsert + strings.Join(inserts, ",") + fmt.Sprintf("RETURNING %s;", returnRows)

	//prepare the statement
	stmt, err := db.Prepare(sqlInsert)
	if err != nil {
		log.Fatalln("ERROR: ", err)
	}
	defer stmt.Close()

	// format all vals at once
	var allInsertedFiles []File
	rows, err := stmt.Query(vals...)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			log.Println(pqErr)
		}
		return nil, err
	}

	for rows.Next() {
		var currentRecord File

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

		currentRecord.FileType = fileType.Dict[fType]
		currentRecord.ObjectType = objectType.Dict[oType]
		currentRecord.ProcessingState = processingState.Dict[pState]
		currentRecord.UploadedState = uploadState.Dict[uState]

		if err != nil {
			log.Println("ERROR: ", err)
		}

		allInsertedFiles = append(allInsertedFiles, currentRecord)

	}

	if err != nil {
		log.Println(err)
	}

	return allInsertedFiles, err
}

// UpdateBucket updates the storage bucket as part of upload process and sets Status
func (p *File) UpdateBucket(db core.PostgresAPI, uploadId string, bucket string, s3Key string, organizationId int64) error {

	queryStr := fmt.Sprintf("UPDATE \"%d\".files SET s3_bucket=$1, s3_key=$2 WHERE UUID=$3;", organizationId)
	result, err := db.Exec(queryStr, bucket, s3Key, uploadId)

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
			nofFoundError := &ErrFileNotFound{}
			log.Println(nofFoundError.Error())
			return nofFoundError
		}

		multipleRowError := &ErrMultipleRowsAffected{}
		log.Println(multipleRowError.Error())
		return multipleRowError
	}

	return nil

}
