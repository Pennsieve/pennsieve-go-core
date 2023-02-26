package dydb

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
