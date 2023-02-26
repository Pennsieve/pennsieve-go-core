package dydb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pennsieve/pennsieve-go-core/pkg/dydb/models"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"
)

const manifestTableName = "upload-table"
const manifestFileTableName = "upload-file-table"

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getDynamoClient() *dynamodb.Client {

	testDBUri := getEnv("DYNAMODB_URL", "http://localhost:8000")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy_secret", "1234")),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: testDBUri}, nil
			})),
	)
	if err != nil {
		panic(err)
	}

	svc := dynamodb.NewFromConfig(cfg)
	return svc
}

func TestMain(m *testing.M) {

	// If testing on Jenkins (-> DYNAMODB_URL is set) then wait for db to be active.
	if _, ok := os.LookupEnv("DYNAMODB_URL"); ok {
		time.Sleep(5 * time.Second)
	}

	svc := getDynamoClient()
	svc.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String("upload-table")})
	svc.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String("upload-file-table")})

	_, err := svc.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ManifestId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("UserId"),
				AttributeType: types.ScalarAttributeTypeN,
			},
			{
				AttributeName: aws.String("DatasetNodeId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ManifestId"),
				KeyType:       types.KeyTypeHash,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("DatasetManifestIndex"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("DatasetNodeId"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("UserId"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					NonKeyAttributes: nil,
					ProjectionType:   "ALL",
				},
				ProvisionedThroughput: nil,
			},
		},
		TableName:   aws.String("upload-table"),
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		log.Printf("Couldn't create table. Here's why: %v\n", err)
	} else {
		waiter := dynamodb.NewTableExistsWaiter(svc)
		err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String("upload-table")}, 5*time.Minute)
		if err != nil {
			log.Printf("Wait for table exists failed. Here's why: %v\n", err)
		}
	}

	_, err = svc.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ManifestId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("UploadId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("Status"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("FilePath"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("InProgress"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ManifestId"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("UploadId"),
				KeyType:       types.KeyTypeRange,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("StatusIndex"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("Status"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("ManifestId"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					NonKeyAttributes: []string{"ManifestId", "UploadId", "FileName", "FilePath", "FileType"},
					ProjectionType:   types.ProjectionTypeInclude,
				},
				ProvisionedThroughput: nil,
			},
			{
				IndexName: aws.String("InProgressIndex"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("ManifestId"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("InProgress"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					NonKeyAttributes: []string{"FileName", "FilePath", "FileType", "Status"},
					ProjectionType:   types.ProjectionTypeInclude,
				},
				ProvisionedThroughput: nil,
			},
			{
				IndexName: aws.String("PathIndex"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("ManifestId"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("FilePath"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					NonKeyAttributes: []string{"FileName", "UploadId", "MergePackageId"},
					ProjectionType:   types.ProjectionTypeInclude,
				},
				ProvisionedThroughput: nil,
			},
		},
		TableName:   aws.String("upload-file-table"),
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		log.Printf("Couldn't create table. Here's why: %v\n", err)
	} else {
		waiter := dynamodb.NewTableExistsWaiter(svc)
		err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String("upload-file-table")}, 5*time.Minute)
		if err != nil {
			log.Printf("Wait for table exists failed. Here's why: %v\n", err)
		}
	}

	// Populate tables
	populateManifestTable()

	// Run tests
	code := m.Run()

	// return
	os.Exit(code)
}

// populateManifestTable populates the test dydb table with entries for testing
func populateManifestTable() {

	ctx := context.Background()

	client := getDynamoClient()
	store := NewDynamoStore(client)

	tb := models.ManifestTable{
		ManifestId:     "1111",
		DatasetId:      1,
		DatasetNodeId:  "N:Dataset:1234",
		OrganizationId: 1,
		UserId:         1,
		Status:         "Unknown",
		DateCreated:    time.Now().Unix(),
	}

	// Create Manifest
	err := store.CreateManifest(ctx, manifestTableName, tb)
	if err != nil {
		log.Fatalln("Unable to create Manifest")
	}

	// Create second upload
	tb2 := models.ManifestTable{
		ManifestId:     "2222",
		DatasetId:      2,
		DatasetNodeId:  "N:Dataset:5678",
		OrganizationId: 1,
		UserId:         1,
		Status:         "Unknown",
		DateCreated:    time.Now().Unix(),
	}

	err = store.CreateManifest(ctx, manifestTableName, tb2)
	if err != nil {
		log.Fatalln("Unable to create Manifest")
	}
	// Create second upload
	tb3 := models.ManifestTable{
		ManifestId:     "3333",
		DatasetId:      2,
		DatasetNodeId:  "N:Dataset:5678",
		OrganizationId: 1,
		UserId:         1,
		Status:         "Unknown",
		DateCreated:    time.Now().Unix(),
	}

	err = store.CreateManifest(ctx, manifestTableName, tb3)
	if err != nil {
		log.Fatalln("Unable to create Manifest")
	}

}
