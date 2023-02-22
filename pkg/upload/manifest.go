package upload

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	core2 "github.com/pennsieve/pennsieve-go-core/pkg/core"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dbTable"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var syncWG sync.WaitGroup

const batchSize = 25 // maximum batch size for batchPut action on dynamodb
const nrWorkers = 2  // preliminary profiling shows that more workers don't improve efficiency for up to 1000 files

type ManifestSession struct {
	FileTableName string
	TableName     string
	Client        core2.DynamoDBAPI
	SNSClient     core2.SnsAPI
	SNSTopic      string
	S3Client      core2.S3API
}

// fileWalk channel used to distribute FileDTOs to the workers importing the files in DynamoDB
type fileWalk chan manifestFile.FileDTO

// CreateManifest creates a new Manifest in DynamoDB
func (s ManifestSession) CreateManifest(item dbTable.ManifestTable) error {

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

	_, err = s.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(s.TableName),
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

// AddFiles manages the workers and defines the go routines to add files to upload db.
func (s ManifestSession) AddFiles(manifestId string, items []manifestFile.FileDTO, forceStatus *manifestFile.Status) *manifest.AddFilesStats {

	// Populate DynamoDB with concurrent workers.
	walker := make(fileWalk, batchSize)
	result := make(chan manifest.AddFilesStats, nrWorkers)

	// List crawler
	go func() {
		// Gather the files to upload by walking the path recursively
		defer func() {
			close(walker)
		}()
		log.WithFields(
			log.Fields{
				"manifest_id": manifestId,
			},
		).Debug(fmt.Sprintf("Adding %d number of items from upload.", len(items)))

		for _, f := range items {
			walker <- f
		}
	}()

	// Initiate a set of upload sync workers as go-routines
	for w := 1; w <= nrWorkers; w++ {
		w2 := int32(w)
		syncWG.Add(1)
		log.Debug("starting worker:", w2)

		go func() {
			stats, _ := s.createOrUpdateFile(w2, walker, manifestId, forceStatus)
			result <- *stats
			defer func() {
				log.Debug("Closing Worker: ", w2)
				syncWG.Done()
			}()
		}()
	}

	syncWG.Wait()
	close(result)

	resp := manifest.AddFilesStats{}
	for r := range result {
		resp.NrFilesUpdated += r.NrFilesUpdated
		resp.NrFilesRemoved += r.NrFilesRemoved
		resp.FileStatus = append(resp.FileStatus, r.FileStatus...)
		resp.FailedFiles = append(resp.FailedFiles, r.FailedFiles...)
	}

	return &resp

}

// updateDynamoDb sends a set of FileDTOs to dynamodb.
func (s ManifestSession) updateDynamoDb(manifestId string, fileSlice []manifestFile.FileDTO, forceStatus *manifestFile.Status) (*manifest.AddFilesStats, error) {
	// Create Batch Put request for the fileslice and update dynamodb with one call
	var writeRequests []types.WriteRequest

	var syncResponses []manifestFile.FileStatusDTO

	// Iterate over files in the fileSlice array and create writeRequests.
	var nrFilesUpdated int
	var nrFilesRemoved int
	for _, file := range fileSlice {
		// Get existing status for file in dynamodb, Unknown if does not exist
		var request *types.WriteRequest
		var setStatus manifestFile.Status
		if forceStatus == nil {
			curStatus, err := s.statusForFileItem(manifestId, &file)
			if err != nil {
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
						"upload_id":   file.UploadID,
					},
				).Error("Unable to check status of existing upload file.")
				return nil, errors.New("unable to check status of existing upload file")
			}

			// Determine the sync action based on provided status and current status.
			request, setStatus, err = s.getAction(manifestId, file, curStatus)
			if err != nil {
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
						"upload_id":   file.UploadID,
					},
				).Error("Unable to get action for upload file.")
				return nil, errors.New("unable to get action for upload file")
			}
		} else {

			isInProgress := forceStatus.IsInProgress()
			item := dbTable.ManifestFileTable{
				ManifestId:     manifestId,
				UploadId:       file.UploadID,
				FilePath:       file.TargetPath,
				FileName:       file.TargetName,
				Status:         forceStatus.String(),
				MergePackageId: file.MergePackageId,
				FileType:       file.FileType,
			}

			if len(isInProgress) > 0 {
				item.InProgress = isInProgress
			}

			data, err := attributevalue.MarshalMap(item)
			if err != nil {
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
						"upload_id":   file.UploadID,
					},
				).Fatalf("MarshalMap: %v\n", err)
			}
			if len(isInProgress) == 0 {
				delete(data, "InProgress")
			}

			request = &types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: data,
				},
			}

			setStatus = *forceStatus
		}

		// If action requires dynamodb actionm add request to array of requests
		if request != nil {
			writeRequests = append(writeRequests, *request)
		}

		// Set the API response object for the file.
		syncResponses = append(syncResponses, manifestFile.FileStatusDTO{
			UploadId: file.UploadID,
			Status:   setStatus,
		})

	}

	var failedFiles []string
	var err error
	if len(writeRequests) > 0 {
		// Format requests and call DynamoDB
		requestItems := map[string][]types.WriteRequest{
			s.FileTableName: writeRequests,
		}

		params := dynamodb.BatchWriteItemInput{
			RequestItems:                requestItems,
			ReturnConsumedCapacity:      "NONE",
			ReturnItemCollectionMetrics: "NONE",
		}

		// Write files to upload file dynamobd table
		data, err := s.Client.BatchWriteItem(context.Background(), &params)
		if err != nil {
			log.WithFields(
				log.Fields{
					"manifest_id": manifestId,
				},
			).Fatalln("Unable to Batch Write: ", err)
		}

		// Handle potential failed files:
		// Step 1: Retry if there are unprocessed files.
		nrRetries := 5
		retryIndex := 0
		unProcessedItems := data.UnprocessedItems
		for len(unProcessedItems) > 0 {
			log.Debug("CONTAINS UNPROCESSED DATA", unProcessedItems)
			params = dynamodb.BatchWriteItemInput{
				RequestItems:                unProcessedItems,
				ReturnConsumedCapacity:      "NONE",
				ReturnItemCollectionMetrics: "NONE",
			}

			data, err = s.Client.BatchWriteItem(context.Background(), &params)
			unProcessedItems = data.UnprocessedItems

			retryIndex++
			if retryIndex == nrRetries {
				log.Warn("Dynamodb did not ingest all the file records.")
				break
			}
			time.Sleep(time.Duration(200*(1+retryIndex)) * time.Millisecond)
		}

		// Step 2: Set the failedFiles array to return failed update to client.
		putRequestList := unProcessedItems[s.FileTableName]
		for _, f := range putRequestList {
			item := f.PutRequest.Item
			fileEntry := dbTable.ManifestFileTable{}
			err = attributevalue.UnmarshalMap(item, &fileEntry)
			if err != nil {
				log.Error("Unable to UnMarshall unprocessed items. ", err)
				return nil, err
			}
			failedFiles = append(failedFiles, fileEntry.UploadId)

			switch fileEntry.Status {
			case manifestFile.Removed.String():
				nrFilesRemoved--
			case manifestFile.Local.String(), manifestFile.Failed.String():
				nrFilesUpdated--
			default:
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
						"upload_id":   fileEntry.UploadId,
					},
				).Warn(fmt.Sprintf("Unable to find status match: %s", fileEntry.Status))
			}
		}
	}

	response := manifest.AddFilesStats{
		NrFilesUpdated: nrFilesUpdated,
		NrFilesRemoved: nrFilesRemoved,
		FileStatus:     syncResponses,
		FailedFiles:    failedFiles,
	}
	return &response, err

}

// createOrUpdateFile is run in a goroutine and grabs set of files from channel and calls updateDynamoDb.
func (s ManifestSession) createOrUpdateFile(workerId int32, files fileWalk, manifestId string, forceStatus *manifestFile.Status) (*manifest.AddFilesStats, error) {

	response := manifest.AddFilesStats{}

	// Create file slice of size "batchSize" or smaller if end of list.
	var fileSlice []manifestFile.FileDTO = nil
	for record := range files {
		fileSlice = append(fileSlice, record)

		// When the number of items in fileSize matches the batchSize --> make call to update dynamodb
		if len(fileSlice) == batchSize {
			stats, _ := s.updateDynamoDb(manifestId, fileSlice, forceStatus)
			fileSlice = nil

			response.NrFilesUpdated += stats.NrFilesUpdated
			response.NrFilesRemoved += stats.NrFilesRemoved
			response.FailedFiles = append(response.FailedFiles, stats.FailedFiles...)
			response.FileStatus = append(response.FileStatus, stats.FileStatus...)
		}
	}

	// Add final partially filled fileSlice to database
	if fileSlice != nil {
		stats, _ := s.updateDynamoDb(manifestId, fileSlice, forceStatus)
		response.NrFilesUpdated += stats.NrFilesUpdated
		response.NrFilesRemoved += stats.NrFilesRemoved
		response.FailedFiles = append(response.FailedFiles, stats.FailedFiles...)
		response.FileStatus = append(response.FileStatus, stats.FileStatus...)
	}

	return &response, nil
}

func (s ManifestSession) statusForFileItem(manifestId string, file *manifestFile.FileDTO) (manifestFile.Status, error) {
	// Get current status in db if exist
	getItemInput := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"ManifestId": &types.AttributeValueMemberS{Value: manifestId},
			"UploadId":   &types.AttributeValueMemberS{Value: file.UploadID},
		},
		TableName: aws.String(s.FileTableName),
	}

	result, err := s.Client.GetItem(context.Background(), getItemInput)
	if err != nil {
		log.Error("Error getting item from dynamodb")
	}

	var pItem dbTable.ManifestFileTable
	if len(result.Item) > 0 {
		err = attributevalue.UnmarshalMap(result.Item, &pItem)
		if err != nil {
			log.Fatalf(err.Error())
		}

		var m manifestFile.Status
		return m.ManifestFileStatusMap(pItem.Status), nil
	}

	return manifestFile.Unknown, nil
}

func (s ManifestSession) getAction(manifestId string, file manifestFile.FileDTO, curStatus manifestFile.Status) (*types.WriteRequest, manifestFile.Status, error) {

	/*
		serverside status: sync, imported, finalized, verified, failed
		clientside status: initiated, sync, imported, verified, failed, unknown

	*/
	item := dbTable.ManifestFileTable{
		ManifestId:     manifestId,
		UploadId:       file.UploadID,
		FilePath:       file.TargetPath,
		FileName:       file.TargetName,
		Status:         manifestFile.Registered.String(),
		MergePackageId: file.MergePackageId,
		FileType:       file.FileType,
	}

	// Switch based on provided status from client
	// file --> provided as part of request
	// curStatus --> current status in dynamodb
	switch file.Status {
	case manifestFile.Removed:
		// File is removed after being synced --> remove from dynamodb if not uploaded already.
		// If uploaded --> return current status

		switch curStatus {
		case manifestFile.Finalized:
			// If client is removed, but server is Finalized --> respond with verified
			// This should never happen but ensures that uploaded files are visible to client.

			item.Status = manifestFile.Verified.String()

			data, err := attributevalue.MarshalMap(item)
			if err != nil {
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
						"upload_id":   file.UploadID,
					},
				).Fatalf("MarshalMap: %v\n", err)
			}
			delete(data, "InProgress")

			request := types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: data,
				},
			}

			return &request, manifestFile.Verified, nil
		case manifestFile.Imported, manifestFile.Verified:
			// If client is removed, but server is Imported/Verified --> respond with server status
			// This should never happen but ensures that uploaded files are visible to client.

			return nil, curStatus, nil
		default:
			// If server synced or failed --> remove from dynamodb
			data, err := attributevalue.MarshalMap(dbTable.ManifestFilePrimaryKey{
				ManifestId: manifestId,
				UploadId:   file.UploadID,
			})
			if err != nil {
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
						"upload_id":   file.UploadID,
					},
				).Fatalf("MarshalMap: %v\n", err)
			}
			request := types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: data,
				},
			}

			return &request, manifestFile.Removed, nil
		}
	case manifestFile.Local, manifestFile.Failed:
		// File is newly created or we are trying to re-upload

		switch curStatus {
		case manifestFile.Finalized:
			// If client is initiated or failed, but server is Finalized --> respond with verified
			// This should never happen but ensures that uploaded files are visible to client.

			item.Status = manifestFile.Verified.String()

			data, err := attributevalue.MarshalMap(item)
			if err != nil {
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
						"upload_id":   file.UploadID,
					},
				).Fatalf("MarshalMap: %v\n", err)
			}
			delete(data, "InProgress")

			request := types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: data,
				},
			}

			return &request, manifestFile.Verified, nil
		case manifestFile.Registered, manifestFile.Failed, manifestFile.Unknown:
			// server is synced, failed, unknown --> add/update the entry in dynamodb
			item.Status = manifestFile.Registered.String()
			item.InProgress = "x"

			data, err := attributevalue.MarshalMap(item)
			if err != nil {
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
						"upload_id":   file.UploadID,
					},
				).Fatalf("MarshalMap: %v\n", err)
			}
			request := types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: data,
				},
			}

			return &request, manifestFile.Registered, nil
		default:
			return nil, curStatus, nil
		}
	case manifestFile.Imported:
		// Last update to file was imported

		switch curStatus {

		case manifestFile.Finalized:
			item.Status = manifestFile.Verified.String()

			data, err := attributevalue.MarshalMap(item)
			if err != nil {
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
						"upload_id":   file.UploadID,
					},
				).Fatalf("MarshalMap: %v\n", err)
			}
			delete(data, "InProgress")

			request := types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: data,
				},
			}

			return &request, manifestFile.Verified, nil
		default:
			return nil, curStatus, nil

		}
	case manifestFile.Registered, manifestFile.Unknown:

		switch curStatus {
		case manifestFile.Registered:
			// server is synced --> update dynamobd in case target path has changed

			item.Status = manifestFile.Registered.String()
			item.InProgress = "x"

			data, err := attributevalue.MarshalMap(item)
			if err != nil {
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
						"upload_id":   file.UploadID,
					},
				).Fatalf("MarshalMap: %v\n", err)
			}

			request := types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: data,
				},
			}

			return &request, manifestFile.Registered, nil
		case manifestFile.Finalized, manifestFile.Imported, manifestFile.Verified:
			// If client is synced and server is Finalized --> respond with verified

			item.Status = curStatus.String()

			data, err := attributevalue.MarshalMap(item)
			if err != nil {
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
						"upload_id":   file.UploadID,
					},
				).Fatalf("MarshalMap: %v\n", err)
			}
			delete(data, "InProgress")

			request := types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: data,
				},
			}

			return &request, manifestFile.Verified, nil
		default:
			return nil, curStatus, nil
		}
	case manifestFile.Finalized, manifestFile.Verified:
		return nil, curStatus, nil

	default:
		return nil, curStatus, nil
	}

	log.WithFields(
		log.Fields{
			"manifest_id": manifestId,
			"upload_id":   file.UploadID,
		},
	).Error("Unhandled case in getAction for file.")
	return nil, manifestFile.Unknown, errors.New("unhandled case in getAction")

}
