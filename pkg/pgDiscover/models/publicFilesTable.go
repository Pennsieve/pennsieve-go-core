package models

import (
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
	"time"
)

type File struct {
	Id              int64         `json:"id"`
	Name            string        `json:"name"`
	DatasetId       int           `json:"dataset_id"`
	Version         int           `json:"version"`
	FileType        fileType.Type `json:"file_type"`
	Size            int64         `json:"size"`
	S3Key           string        `json:"s3_key"`
	Path            string        `json:"path"`
	SourcePackageId int64         `json:"source_package_id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}
