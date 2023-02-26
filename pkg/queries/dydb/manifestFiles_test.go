package dydb

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManifestFileStore(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, client *DynamoStore,
	){
		"get writeRequests based on status":             testGetAction,
		"test removing failed files from sync response": testRemoveFailedFilesFromResponse,
		"add files to manifest":                         testSyncFiles,
	} {
		t.Run(scenario, func(t *testing.T) {
			client := getDynamoClient()
			store := NewDynamoStore(client)
			fn(t, store)
		})
	}
}

func testSyncFiles(t *testing.T, client *DynamoStore) {
	manifestId := "1"
	dtos := []manifestFile.FileDTO{
		{
			UploadID:   "1",
			TargetName: "test1",
			Status:     manifestFile.Local,
		},
		{
			UploadID:   "2",
			TargetName: "test2",
			Status:     manifestFile.Local,
		},
		{
			UploadID:   "3",
			TargetName: "test3",
			Status:     manifestFile.Local,
		},
	}

	ctx := context.Background()
	stats, err := client.SyncFiles(ctx, manifestFileTableName, manifestId, dtos, nil)
	assert.Nil(t, err, "Manifest files could not be added")
	assert.Equal(t, 3, stats.NrFilesUpdated, "Number of files updated does not match")

	out, err := client.GetManifestFile(ctx, manifestFileTableName, manifestId, "1")
	assert.Nil(t, err, "Manifest file could not be retrieved")
	assert.Equal(t, "1", out.ManifestId)
	assert.Equal(t, "test1", out.FileName)
	t.Log(out)

}

func testRemoveFailedFilesFromResponse(t *testing.T, client *DynamoStore) {
	syncResp := []manifestFile.FileStatusDTO{
		{
			UploadId: "1",
			Status:   manifestFile.Registered,
		},
		{
			UploadId: "2",
			Status:   manifestFile.Registered,
		},
		{
			UploadId: "3",
			Status:   manifestFile.Registered,
		},
		{
			UploadId: "4",
			Status:   manifestFile.Registered,
		},
	}

	failedFiles := []string{"2", "4"}

	resp := removeFailedFilesFromResponse(failedFiles, syncResp)
	assert.Equal(t, resp, []manifestFile.FileStatusDTO{
		{
			UploadId: "1",
			Status:   manifestFile.Registered,
		},
		{
			UploadId: "3",
			Status:   manifestFile.Registered,
		},
	})
}

func testGetAction(t *testing.T, client *DynamoStore) {
	manifestId := "getActionTest"

	file := manifestFile.FileDTO{
		UploadID: "getActionTestId",
		Status:   manifestFile.Local,
	}

	// Check file that is newly uploaded and not in manifest Local (Unknown) --> Registered
	req, status, err := GetWriteRequest(manifestId, file, manifestFile.Unknown)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Registered, status)
	assert.NotNil(t, req.PutRequest)

	// Check file that is removed locally and previously registered: Removed (Registered) --> Delete request
	file.Status = manifestFile.Removed
	req, status, err = GetWriteRequest(manifestId, file, manifestFile.Registered)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Removed, status)
	assert.NotNil(t, req.DeleteRequest)

	// Failed (Failed) --> Registered
	file.Status = manifestFile.Failed
	req, status, err = GetWriteRequest(manifestId, file, manifestFile.Registered)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Registered, status)
	assert.NotNil(t, req.PutRequest)

	// Imported (Finalized) --> Finalized
	file.Status = manifestFile.Imported
	req, status, err = GetWriteRequest(manifestId, file, manifestFile.Finalized)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Verified, status)
	assert.NotNil(t, req.PutRequest)

	// Imported (Imported) --> Imported
	file.Status = manifestFile.Imported
	req, status, err = GetWriteRequest(manifestId, file, manifestFile.Imported)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Imported, status)
	assert.Nil(t, req)

	// Registered (Registered) --> Registered
	file.Status = manifestFile.Registered
	req, status, err = GetWriteRequest(manifestId, file, manifestFile.Registered)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Registered, status)
	assert.NotNil(t, req.PutRequest)

	// Registered (Finalized) --> Verified
	file.Status = manifestFile.Registered
	req, status, err = GetWriteRequest(manifestId, file, manifestFile.Finalized)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Verified, status)
	assert.NotNil(t, req.PutRequest)

	// Registered (Imported) --> Verified
	file.Status = manifestFile.Registered
	req, status, err = GetWriteRequest(manifestId, file, manifestFile.Imported)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Verified, status)
	assert.NotNil(t, req.PutRequest)

	// Registered (Verified) --> Verified
	file.Status = manifestFile.Registered
	req, status, err = GetWriteRequest(manifestId, file, manifestFile.Verified)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Verified, status)
	assert.NotNil(t, req.PutRequest)

	// Finalized (Finalized) --> Finalized
	file.Status = manifestFile.Finalized
	req, status, err = GetWriteRequest(manifestId, file, manifestFile.Finalized)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Finalized, status)
	assert.Nil(t, req)

	// Verified (Verified) --> Verified
	file.Status = manifestFile.Verified
	req, status, err = GetWriteRequest(manifestId, file, manifestFile.Verified)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Verified, status)
	assert.Nil(t, req)

	// Unknown (Registered) --> Registered
	file.Status = manifestFile.Registered
	req, status, err = GetWriteRequest(manifestId, file, manifestFile.Registered)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Registered, status)
	assert.NotNil(t, req.PutRequest)

}
