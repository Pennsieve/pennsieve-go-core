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
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// PUBLIC METHODS

type fileWalk chan manifestFile.FileDTO

var syncWG sync.WaitGroup

const batchSize = 25 // maximum batch size for batchPut action on dydb
const nrWorkers = 2  // preliminary profiling shows that more workers don't improve efficiency for up to 1000 files

// SyncFiles adds or updates files in the manifest file table
func (q *Queries) SyncFiles(ctx context.Context, tableName string, manifestId string, fileSlice []manifestFile.FileDTO, forceStatus *manifestFile.Status) (*manifest.AddFilesStats, error) {
	// Create Batch Put request for the fileslice and update dydb with one call
	var writeRequests []types.WriteRequest

	var syncResponses []manifestFile.FileStatusDTO

	// Iterate over files in the fileSlice array and create writeRequests.
	var nrFilesUpdated int
	var nrFilesRemoved int
	for _, file := range fileSlice {
		// Get existing status for file in dydb, Unknown if does not exist
		var request *types.WriteRequest
		var setStatus manifestFile.Status
		if forceStatus == nil {
			curStatus, err := q.statusForFileItem(ctx, tableName, manifestId, &file)
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
			request, setStatus, err = GetWriteRequest(manifestId, file, curStatus)
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
			item := dydb.ManifestFileTable{
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

		// If action requires dydb actionm add request to array of requests
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
			tableName: writeRequests,
		}

		params := dynamodb.BatchWriteItemInput{
			RequestItems:                requestItems,
			ReturnConsumedCapacity:      "NONE",
			ReturnItemCollectionMetrics: "NONE",
		}

		// Write files to upload file dynamobd table
		data, err := q.db.BatchWriteItem(ctx, &params)
		if err != nil {
			log.WithFields(
				log.Fields{
					"manifest_id": manifestId,
				},
			).Fatalln("Unable to Batch Write: ", err)
		}

		nrFilesUpdated += len(writeRequests) - len(data.UnprocessedItems)

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

			data, err = q.db.BatchWriteItem(context.Background(), &params)
			if err != nil {
				log.WithFields(
					log.Fields{
						"manifest_id": manifestId,
					},
				).Fatalln("Unable to Batch Write: ", err)
			}

			nrFilesUpdated += len(unProcessedItems) - len(data.UnprocessedItems)

			unProcessedItems = data.UnprocessedItems

			retryIndex++
			if retryIndex == nrRetries {
				log.Warn("Dynamodb did not ingest all the file records.")
				break
			}
			time.Sleep(time.Duration(200*(1+retryIndex)) * time.Millisecond)
		}

		// Step 2: Set the failedFiles array to return failed update to client.
		if len(unProcessedItems) > 0 {
			// Create list of uploadIds that failed to be created in table
			putRequestList := unProcessedItems[tableName]
			for _, f := range putRequestList {
				item := f.PutRequest.Item
				fileEntry := dydb.ManifestFileTable{}
				err = attributevalue.UnmarshalMap(item, &fileEntry)
				if err != nil {
					log.Error("Unable to UnMarshall unprocessed items. ", err)
					return nil, err
				}
				failedFiles = append(failedFiles, fileEntry.UploadId)
			}

			// Remove failed files from syncResponse
			syncResponses = removeFailedFilesFromResponse(failedFiles, syncResponses)
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

// UpdateFileTableStatus updates the status of the file in the file-table dydb
func (q *Queries) UpdateFileTableStatus(ctx context.Context, tableName string, manifestId string, uploadId string, status manifestFile.Status, msg string) error {

	_, err := q.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
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
func (q *Queries) GetFilesForPath(ctx context.Context, tableName string, manifestId string, path string, filter string,
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

	result, err := q.db.Query(ctx, &queryInput)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetManifestFile returns a upload file from the ManifestFile Table.
func (q *Queries) GetManifestFile(ctx context.Context, tableName string, manifestId string, uploadId string) (*dydb.ManifestFileTable, error) {
	item := dydb.ManifestFileTable{}

	data, err := q.db.GetItem(ctx, &dynamodb.GetItemInput{
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
func (q *Queries) GetFilesPaginated(ctx context.Context, tableName string, manifestId string, status sql.NullString,
	limit int32, startKey map[string]types.AttributeValue) ([]dydb.ManifestFileTable, map[string]types.AttributeValue, error) {

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
		// Query from main dydb
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

	result, err := q.db.Query(ctx, &queryInput)
	if err != nil {
		return nil, nil, err
	}

	var items []dydb.ManifestFileTable
	for _, item := range result.Items {
		fmt.Println("Hello item: ", item)
		manifestFile := dydb.ManifestFileTable{}
		err = attributevalue.UnmarshalMap(item, &manifestFile)
		if err != nil {
			return nil, nil, fmt.Errorf("UnmarshalMap: %v\n", err)
		}
		items = append(items, manifestFile)
	}

	return items, result.LastEvaluatedKey, nil
}

// AddFiles manages the workers and defines the go routines to add files to upload db.
func (q *Queries) AddFiles(manifestId string, items []manifestFile.FileDTO, forceStatus *manifestFile.Status, fileNameTable string) *manifest.AddFilesStats {

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
			stats, _ := q.createOrUpdateFile(w2, walker, manifestId, forceStatus, fileNameTable)
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

// PRIVATE METHODS
func (q *Queries) statusForFileItem(ctx context.Context, tableName string, manifestId string, file *manifestFile.FileDTO) (manifestFile.Status, error) {
	// Get current status in db if exist
	getItemInput := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"ManifestId": &types.AttributeValueMemberS{Value: manifestId},
			"UploadId":   &types.AttributeValueMemberS{Value: file.UploadID},
		},
		TableName: aws.String(tableName),
	}

	result, err := q.db.GetItem(context.Background(), getItemInput)
	if err != nil {
		log.Error("Error getting item from dydb")
	}

	var pItem dydb.ManifestFileTable
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

// createOrUpdateFile is run in a goroutine and grabs set of files from channel and calls updateDynamoDb.
func (q *Queries) createOrUpdateFile(workerId int32, files fileWalk, manifestId string, forceStatus *manifestFile.Status, fileTableName string) (*manifest.AddFilesStats, error) {

	//store := dydb.NewDynamoStore(s.Client)
	ctx := context.Background()

	response := manifest.AddFilesStats{}

	// Create file slice of size "batchSize" or smaller if end of list.
	var fileSlice []manifestFile.FileDTO = nil
	for record := range files {
		fileSlice = append(fileSlice, record)

		// When the number of items in fileSize matches the batchSize --> make call to update dydb
		if len(fileSlice) == batchSize {
			stats, _ := q.SyncFiles(ctx, fileTableName, manifestId, fileSlice, forceStatus)
			fileSlice = nil

			response.NrFilesUpdated += stats.NrFilesUpdated
			response.NrFilesRemoved += stats.NrFilesRemoved
			response.FailedFiles = append(response.FailedFiles, stats.FailedFiles...)
			response.FileStatus = append(response.FileStatus, stats.FileStatus...)
		}
	}

	// Add final partially filled fileSlice to database
	if fileSlice != nil {
		stats, _ := q.SyncFiles(ctx, fileTableName, manifestId, fileSlice, forceStatus)
		response.NrFilesUpdated += stats.NrFilesUpdated
		response.NrFilesRemoved += stats.NrFilesRemoved
		response.FailedFiles = append(response.FailedFiles, stats.FailedFiles...)
		response.FileStatus = append(response.FileStatus, stats.FileStatus...)
	}

	return &response, nil
}

// getAction returns the writeRequests for a given fileDTO and current status
func GetWriteRequest(manifestId string, file manifestFile.FileDTO, curStatus manifestFile.Status) (*types.WriteRequest, manifestFile.Status, error) {

	/*
		serverside status: sync, imported, finalized, verified, failed
		clientside status: initiated, sync, imported, verified, failed, unknown

	*/
	item := dydb.ManifestFileTable{
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
	// curStatus --> current status in dydb
	switch file.Status {
	case manifestFile.Removed:
		// File is removed after being synced --> remove from dydb if not uploaded already.
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
			// If server synced or failed --> remove from dydb
			data, err := attributevalue.MarshalMap(dydb.ManifestFilePrimaryKey{
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
			// server is synced, failed, unknown --> add/update the entry in dydb
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

// RemoveFailedFilesFromResponse removes any files from the syncResponse that has not been successfully created in dydb
func removeFailedFilesFromResponse(failedRequests []string, syncResponses []manifestFile.FileStatusDTO) []manifestFile.FileStatusDTO {

	var newResponses []manifestFile.FileStatusDTO
	for _, item := range syncResponses {
		if !stringInSlice(item.UploadId, failedRequests) {
			newResponses = append(newResponses, item)
		}
	}
	return newResponses
}

// stringInSlice checks if a string is present in an array of strings
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
