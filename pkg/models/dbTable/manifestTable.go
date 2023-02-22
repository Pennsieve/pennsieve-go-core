package dbTable

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pennsieve/pennsieve-go-core/pkg/core"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest"
	log "github.com/sirupsen/logrus"
)

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

// CreateManifest creates a new Manifest in DynamoDB
func (m *ManifestTable) CreateManifest(client core.DynamoDBAPI, manifestTableName string, item ManifestTable) error {

	data, err := attributevalue.MarshalMap(item)
	if err != nil {
		log.WithFields(
			log.Fields{
				"organization_id": item.OrganizationId,
				"dataset_id":      item.DatasetId,
				"manifest_id":     item.ManifestId,
				"user_id":         item.UserId,
			},
		).Error(fmt.Sprintf("MarshalMap: %v\n", err))
		return fmt.Errorf("MarshalMap: %v\n", err)
	}

	_, err = client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(manifestTableName),
		Item:      data,
	})

	if err != nil {
		log.WithFields(
			log.Fields{
				"organization_id": item.OrganizationId,
				"dataset_id":      item.DatasetId,
				"manifest_id":     item.ManifestId,
				"user_id":         item.UserId,
			},
		).Error(fmt.Sprintf("Error creating upload: %v\n", err))
		return errors.New("Error creating Manifest")
	}

	return nil
}

// GetFromManifest returns a Manifest item for a given upload ID.
func (m *ManifestTable) GetFromManifest(client core.DynamoDBAPI, manifestTableName string, manifestId string) (*ManifestTable, error) {

	item := ManifestTable{}

	data, err := client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(manifestTableName),
		Key: map[string]types.AttributeValue{
			"ManifestId": &types.AttributeValueMemberS{Value: manifestId},
		},
	})

	if err != nil {
		return &item, fmt.Errorf("GetItem: %v\n", err)
	}

	if data.Item == nil {
		return &item, fmt.Errorf("GetItem: Manifest not found.\n")
	}

	err = attributevalue.UnmarshalMap(data.Item, &item)
	if err != nil {
		return &item, fmt.Errorf("UnmarshalMap: %v\n", err)
	}

	return &item, nil
}

// GetManifestsForDataset returns all manifests for a given dataset.
func (m *ManifestTable) GetManifestsForDataset(client core.DynamoDBAPI, manifestTableName string, datasetNodeId string) ([]ManifestTable, error) {

	queryInput := dynamodb.QueryInput{
		TableName:              aws.String(manifestTableName),
		IndexName:              aws.String("DatasetManifestIndex"),
		KeyConditionExpression: aws.String("DatasetNodeId = :datasetValue"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":datasetValue": &types.AttributeValueMemberS{Value: datasetNodeId},
		},
		Select: "ALL_ATTRIBUTES",
	}

	result, err := client.Query(context.Background(), &queryInput)
	if err != nil {
		return nil, err
	}

	items := []ManifestTable{}
	for _, item := range result.Items {
		manifest := ManifestTable{}
		err = attributevalue.UnmarshalMap(item, &manifest)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalMap: %v\n", err)
		}
		items = append(items, manifest)
	}

	return items, nil
}

// UpdateManifestStatus updates the status of the upload in dynamodb
func (m *ManifestTable) UpdateManifestStatus(client core.DynamoDBAPI, tableName string, manifestId string, status manifest.Status) error {

	_, err := client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ManifestId": &types.AttributeValueMemberS{Value: manifestId},
		},
		UpdateExpression: aws.String("set #status = :statusValue"),
		ExpressionAttributeNames: map[string]string{
			"#status": "Status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":statusValue": &types.AttributeValueMemberS{Value: status.String()},
		},
	})
	return err
}
