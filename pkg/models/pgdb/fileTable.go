package pgdb

import (
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/objectType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/processingState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/uploadState"
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
