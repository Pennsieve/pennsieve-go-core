package dydb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DynamoStore provides the Queries interface and a db instance.
type DynamoStore struct {
	*Queries
	db *dynamodb.Client
}

// NewDynamoStore returns a SQLStore object which implements the Queires
func NewDynamoStore(db *dynamodb.Client) *DynamoStore {
	return &DynamoStore{
		db:      db,
		Queries: New(db),
	}
}
