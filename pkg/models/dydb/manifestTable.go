package dydb

// ManifestTable is a representation of a Manifest in DynamoDB
type ManifestTable struct {
	ManifestId     string `dynamodbav:"ManifestId"`
	DatasetId      int64  `dynamodbav:"DatasetId"`
	DatasetNodeId  string `dynamodbav:"DatasetNodeId"`
	OrganizationId int64  `dynamodbav:"OrganizationId"`
	UserId         int64  `dynamodbav:"UserId"`
	Status         string `dynamodbav:"Status"`
	DateCreated    int64  `dynamodbav:"DateCreated"`
}
