package dydb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// From https://dev.to/techschoolguru/a-clean-way-to-implement-database-transaction-in-golang-2ba

// DB Default interface with methods that are available for both DB adn TX sessions.
type DB interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	BatchExecuteStatement(ctx context.Context, params *dynamodb.BatchExecuteStatementInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchExecuteStatementOutput, error)
}

// New returns a Queries object backed by a DB interface (either DB or TX)
func New(db DB) *Queries {
	return &Queries{db: db}
}

// Queries is a struct with a db object that implements the DB interface.
// This means that db can either be a direct DB connection or a TX transaction.
type Queries struct {
	db DB
}
