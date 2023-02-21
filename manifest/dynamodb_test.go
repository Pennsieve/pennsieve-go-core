package manifest

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pennsieve/pennsieve-go-core/models/dbTable"
	"github.com/pennsieve/pennsieve-go-core/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/models/manifest/manifestFile"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getClient() *dynamodb.Client {

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

	svc := getClient()
	svc.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String("manifest-table")})
	svc.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String("manifest-file-table")})

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
			{
				AttributeName: aws.String("UserId"),
				KeyType:       types.KeyTypeRange,
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
		TableName:   aws.String("manifest-table"),
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		log.Printf("Couldn't create table. Here's why: %v\n", err)
	} else {
		waiter := dynamodb.NewTableExistsWaiter(svc)
		err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String("manifest-table")}, 5*time.Minute)
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
		TableName:   aws.String("manifest-file-table"),
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		log.Printf("Couldn't create table. Here's why: %v\n", err)
	} else {
		waiter := dynamodb.NewTableExistsWaiter(svc)
		err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String("manifest-file-table")}, 5*time.Minute)
		if err != nil {
			log.Printf("Wait for table exists failed. Here's why: %v\n", err)
		}
	}

	// Run tests
	code := m.Run()

	// return
	os.Exit(code)
}

func TestManifest(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, ms *ManifestSession,
	){
		"create and get manifest": testCreateGetManifest,
		"Add files to manifest":   testAddFiles,
	} {
		t.Run(scenario, func(t *testing.T) {
			client := getClient()
			ms := ManifestSession{
				FileTableName: "manifest-file-table",
				TableName:     "manifest-table",
				Client:        client,
				SNSClient:     nil,
				SNSTopic:      "",
				S3Client:      nil,
			}

			fn(t, &ms)
		})
	}
}

func testCreateGetManifest(t *testing.T, ms *ManifestSession) {

	tb := dbTable.ManifestTable{
		ManifestId:     "1111",
		DatasetId:      1,
		DatasetNodeId:  "N:Dataset:1234",
		OrganizationId: 1,
		UserId:         1,
		Status:         "Unknown",
		DateCreated:    time.Now().Unix(),
	}

	// Create Manifest
	err := ms.CreateManifest(tb)
	assert.Nil(t, err, "Manifest 1 could not be created")

	// Create second manifest
	tb2 := dbTable.ManifestTable{
		ManifestId:     "2222",
		DatasetId:      2,
		DatasetNodeId:  "N:Dataset:5678",
		OrganizationId: 1,
		UserId:         1,
		Status:         "Unknown",
		DateCreated:    time.Now().Unix(),
	}

	err = ms.CreateManifest(tb2)
	assert.Nil(t, err, "Manifest 2 could not be created")

	// Create second manifest
	tb3 := dbTable.ManifestTable{
		ManifestId:     "3333",
		DatasetId:      2,
		DatasetNodeId:  "N:Dataset:5678",
		OrganizationId: 1,
		UserId:         1,
		Status:         "Unknown",
		DateCreated:    time.Now().Unix(),
	}

	err = ms.CreateManifest(tb3)
	assert.Nil(t, err, "Manifest 3 could not be created")

	// Get Manifest
	out, err := dbTable.GetManifestsForDataset(ms.Client, "manifest-table", "N:Dataset:1234")
	assert.Nil(t, err, "Manifest could not be fetched")
	assert.Equal(t, 1, len(out))
	assert.Equal(t, "1111", out[0].ManifestId)
	assert.Equal(t, int64(1), out[0].OrganizationId)
	assert.Equal(t, int64(1), out[0].UserId)

	// Check that there are two manifests for N:Dataset:5678
	out, err = dbTable.GetManifestsForDataset(ms.Client, "manifest-table", "N:Dataset:5678")
	assert.Nil(t, err, "Manifest could not be fetched")
	assert.Equal(t, 2, len(out))
	assert.Equal(t, "2222", out[0].ManifestId)
	assert.Equal(t, "3333", out[1].ManifestId)
}

func testAddFiles(t *testing.T, ms *ManifestSession) {

	testFileDTOs := []manifestFile.FileDTO{
		{
			UploadID:       "111",
			S3Key:          "",
			TargetPath:     "folder1",
			TargetName:     "file1",
			Status:         manifestFile.Unknown,
			MergePackageId: "",
			FileType:       fileType.Aperio.String(),
		},
		{
			UploadID:       "222",
			S3Key:          "",
			TargetPath:     "folder1",
			TargetName:     "file2",
			Status:         manifestFile.Unknown,
			MergePackageId: "",
			FileType:       fileType.Aperio.String(),
		},
	}
	manifestId := "1111"

	result := ms.AddFiles(manifestId, testFileDTOs, nil)
	assert.Equal(t, manifestFile.Local, result.FileStatus[0].Status)

}

//func testGetAction(t *testing.T, svc *dynamodb.Client) {
//
//	getAction(manifestId string, file manifestFile.FileDTO, curStatus manifestFile.Status)
//
//}
