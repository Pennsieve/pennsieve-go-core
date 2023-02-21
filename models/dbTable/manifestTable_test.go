package dbTable

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/middleware"
	"github.com/pennsieve/pennsieve-go-core/core"
)

type MockDynamoDBClient struct {
	core.DynamoDBAPI
}

func (c MockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return &dynamodb.GetItemOutput{
		ConsumedCapacity: nil,
		Item:             nil,
		ResultMetadata:   middleware.Metadata{},
	}, nil
}
