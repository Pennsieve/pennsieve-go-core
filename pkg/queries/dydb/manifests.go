package dydb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dydb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest"
	log "github.com/sirupsen/logrus"
)

// CreateManifest creates a new Manifest in DynamoDB
func (q *Queries) CreateManifest(ctx context.Context, manifestTableName string, item dydb.ManifestTable) error {

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

	getRequest := dynamodb.GetItemInput{
		Key:       data,
		TableName: aws.String("manifestTableName"),
	}
	result, _ := q.db.GetItem(ctx, &getRequest)
	if result != nil {
		return errors.New("manifest with provided ID already exists")
	}

	_, err = q.db.PutItem(context.TODO(), &dynamodb.PutItemInput{
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
		return errors.New("error creating Manifest")
	}

	return nil
}

// GetManifestById returns a Manifest item for a given manifest ID.
func (q *Queries) GetManifestById(ctx context.Context, manifestTableName string, manifestId string) (*dydb.ManifestTable, error) {

	item := dydb.ManifestTable{}

	data, err := q.db.GetItem(ctx, &dynamodb.GetItemInput{
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
func (q *Queries) GetManifestsForDataset(ctx context.Context, manifestTableName string,
	datasetNodeId string) ([]dydb.ManifestTable, error) {

	queryInput := dynamodb.QueryInput{
		TableName:              aws.String(manifestTableName),
		IndexName:              aws.String("DatasetManifestIndex"),
		KeyConditionExpression: aws.String("DatasetNodeId = :datasetValue"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":datasetValue": &types.AttributeValueMemberS{Value: datasetNodeId},
		},
		Select: "ALL_ATTRIBUTES",
	}

	result, err := q.db.Query(context.Background(), &queryInput)
	if err != nil {
		return nil, err
	}

	items := []dydb.ManifestTable{}
	for _, item := range result.Items {
		manifest := dydb.ManifestTable{}
		err = attributevalue.UnmarshalMap(item, &manifest)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalMap: %v\n", err)
		}
		items = append(items, manifest)
	}

	return items, nil
}

// CheckUpdateManifestStatus checks current status of Manifest and updates if necessary.
func (q *Queries) CheckUpdateManifestStatus(ctx context.Context, manifestFileTableName string, manifestTableName string,
	manifestId string, currentStatus string) (manifest.Status, error) {

	// Check if there are any remaining items for manifest and
	// set manifest status if not
	reqStatus := sql.NullString{
		String: "InProgress",
		Valid:  true,
	}

	setStatus := manifest.Initiated

	remaining, _, err := q.GetFilesPaginated(ctx, manifestFileTableName,
		manifestId, reqStatus, 1, nil)
	if err != nil {
		return setStatus, err
	}

	if len(remaining) == 0 {
		setStatus = manifest.Completed
		err = q.updateManifestStatus(ctx, manifestTableName, manifestId, setStatus)
		if err != nil {
			return setStatus, err
		}
	} else if currentStatus == "Completed" {
		setStatus = manifest.Uploading
		err = q.updateManifestStatus(ctx, manifestTableName, manifestId, setStatus)
		if err != nil {
			return setStatus, err
		}
	}

	return setStatus, nil

}

// UpdateManifestStatus updates the status of the upload in dydb
func (q *Queries) updateManifestStatus(ctx context.Context, tableName string, manifestId string, status manifest.Status) error {

	_, err := q.db.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
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
