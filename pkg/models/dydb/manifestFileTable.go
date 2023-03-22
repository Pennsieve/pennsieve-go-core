package dydb

import "reflect"

// ManifestFileTable is a representation of a ManifestFile in DynamoDB
type ManifestFileTable struct {
	ManifestId     string `dynamodbav:"ManifestId"`
	UploadId       string `dynamodbav:"UploadId"`
	FilePath       string `dynamodbav:"FilePath,omitempty"`
	FileName       string `dynamodbav:"FileName"`
	MergePackageId string `dynamodbav:"MergePackageId,omitempty"`
	Status         string `dynamodbav:"Status"`
	FileType       string `dynamodbav:"FileType"`
	InProgress     string `dynamodbav:"InProgress"`
}

type ManifestFilePrimaryKey struct {
	ManifestId string `dynamodbav:"ManifestId"`
	UploadId   string `dynamodbav:"UploadId"`
}

func (m ManifestFileTable) GetHeaders() []string {
	t := reflect.TypeOf(m)

	header := make([]string, t.NumField())
	for i := range header {
		header[i] = t.Field(i).Name
	}

	return header
}

func (m ManifestFileTable) ToSlice() []string {
	return []string{
		m.ManifestId,
		m.UploadId,
		m.FilePath,
		m.FileName,
		m.MergePackageId,
		m.Status,
		m.FileType,
		m.InProgress,
	}
}
