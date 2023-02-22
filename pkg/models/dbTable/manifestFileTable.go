package dbTable

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pennsieve/pennsieve-go-core/pkg/core"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
)

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

// UpdateFileTableStatus updates the status of the file in the file-table dynamodb
func UpdateFileTableStatus(client core.DynamoDBAPI, tableName string, manifestId string, uploadId string, status manifestFile.Status, msg string) error {

	_, err := client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ManifestId": &types.AttributeValueMemberS{Value: manifestId},
			"UploadId":   &types.AttributeValueMemberS{Value: uploadId},
		},
		UpdateExpression: aws.String("set #status = :statusValue, #msg = :msgValue"),
		ExpressionAttributeNames: map[string]string{
			"#status": "Status",
			"#msg":    "Message",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":statusValue": &types.AttributeValueMemberS{Value: status.String()},
			":msgValue":    &types.AttributeValueMemberS{Value: msg},
		},
	})
	return err
}

// GetFilesForPath returns files in path for a upload with optional filter.
func GetFilesForPath(client core.DynamoDBAPI, tableName string, manifestId string, path string, filter string,
	limit int32, startKey map[string]types.AttributeValue) (*dynamodb.QueryOutput, error) {

	queryInput := dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("PathIndex"),
		ExclusiveStartKey:         startKey,
		ExpressionAttributeNames:  nil,
		ExpressionAttributeValues: nil,
		FilterExpression:          aws.String(filter),
		KeyConditionExpression:    aws.String(fmt.Sprintf("partitionKeyName=%s AND sortKeyName=%s", manifestId, path)),
		Limit:                     &limit,
		Select:                    "ALL_ATTRIBUTES",
	}

	result, err := client.Query(context.Background(), &queryInput)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetManifestFile returns a upload file from the ManifestFile Table.
func GetManifestFile(client core.DynamoDBAPI, tableName string, manifestId string, uploadId string) (*ManifestFileTable, error) {
	item := ManifestFileTable{}

	data, err := client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ManifestId": &types.AttributeValueMemberS{Value: manifestId},
			"UploadId":   &types.AttributeValueMemberS{Value: uploadId},
		},
	})

	if err != nil {
		return &item, fmt.Errorf("GetItem: %v\n", err)
	}

	if data.Item == nil {
		return &item, fmt.Errorf("GetItem: ManifestFile not found.\n")
	}

	err = attributevalue.UnmarshalMap(data.Item, &item)
	if err != nil {
		return &item, fmt.Errorf("UnmarshalMap: %v\n", err)
	}

	return &item, nil
}

// GetFilesPaginated returns paginated list of files for a given manifestID and optional status.
func GetFilesPaginated(client core.DynamoDBAPI, tableName string, manifestId string, status sql.NullString,
	limit int32, startKey map[string]types.AttributeValue) ([]ManifestFileTable, map[string]types.AttributeValue, error) {

	var queryInput dynamodb.QueryInput
	switch status.Valid {
	case true:
		if status.String == "InProgress" {
			// Query from Status index
			queryInput = dynamodb.QueryInput{
				TableName:              aws.String(tableName),
				IndexName:              aws.String("InProgressIndex"),
				ExclusiveStartKey:      startKey,
				KeyConditionExpression: aws.String("ManifestId = :manifestValue"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":manifestValue": &types.AttributeValueMemberS{Value: manifestId},
				},
				Limit:  &limit,
				Select: "ALL_PROJECTED_ATTRIBUTES",
			}
		} else {
			// Query from Status index
			queryInput = dynamodb.QueryInput{
				TableName:         aws.String(tableName),
				IndexName:         aws.String("StatusIndex"),
				ExclusiveStartKey: startKey,
				ExpressionAttributeNames: map[string]string{
					"#S": "Status",
				},
				KeyConditionExpression: aws.String("ManifestId = :manifestValue AND #S = :statusValue"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":manifestValue": &types.AttributeValueMemberS{Value: manifestId},
					":statusValue":   &types.AttributeValueMemberS{Value: status.String},
				},
				Limit:  &limit,
				Select: "ALL_PROJECTED_ATTRIBUTES",
			}
		}
	case false:
		// Query from main dynamodb
		queryInput = dynamodb.QueryInput{
			TableName:              aws.String(tableName),
			ExclusiveStartKey:      startKey,
			KeyConditionExpression: aws.String("ManifestId = :manifestValue"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":manifestValue": &types.AttributeValueMemberS{Value: manifestId},
			},
			Limit:  &limit,
			Select: "ALL_ATTRIBUTES",
		}
	}

	result, err := client.Query(context.Background(), &queryInput)
	if err != nil {
		return nil, nil, err
	}

	var items []ManifestFileTable
	for _, item := range result.Items {
		fmt.Println("Hello item: ", item)
		manifestFile := ManifestFileTable{}
		err = attributevalue.UnmarshalMap(item, &manifestFile)
		if err != nil {
			return nil, nil, fmt.Errorf("UnmarshalMap: %v\n", err)
		}
		items = append(items, manifestFile)
	}

	return items, result.LastEvaluatedKey, nil
}
