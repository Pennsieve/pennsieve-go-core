package dydb

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dydb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
	manifestId := "0001"
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
	//
	ctx := context.Background()
	err := client.CreateManifest(ctx, manifestTableName, dydb.ManifestTable{
		ManifestId:     manifestId,
		DatasetId:      1,
		DatasetNodeId:  "N:Dataset:0001",
		OrganizationId: 1,
		UserId:         1,
		Status:         manifest.Initiated.String(),
		DateCreated:    time.Now().Unix(),
	})
	assert.NoError(t, err)
	//
	stats := client.SyncFiles(manifestId, dtos, nil, manifestTableName, manifestFileTableName)
	//assert.Nil(t, err, "Manifest files could not be added")
	assert.Equal(t, 3, stats.NrFilesUpdated, "Number of files updated does not match")

	out, err := client.GetManifestFile(ctx, manifestFileTableName, manifestId, "1")
	assert.Nil(t, err, "Manifest file could not be retrieved")
	assert.Equal(t, "0001", out.ManifestId)
	assert.Equal(t, "test1", out.FileName)

	// Get Manifest and check status
	m, err := client.GetManifestById(ctx, manifestTableName, manifestId)
	assert.NoError(t, err)
	assert.Equal(t, manifest.Initiated.String(), m.Status, "Check 1: Manifest status should be initialized")

	// Check Status
	_, err = client.CheckUpdateManifestStatus(ctx, manifestFileTableName, manifestTableName, manifestId, m.Status)
	assert.NoError(t, err)

	// Check status should not have impacted manifest status
	m, err = client.GetManifestById(ctx, manifestTableName, manifestId)
	assert.NoError(t, err)
	assert.Equal(t, manifest.Initiated.String(), m.Status, "Check 2: Manifest status should be initialized")

	// SET TO UPLOADED --> Manifest status should remain in progress
	for _, f := range dtos {
		err := client.UpdateFileTableStatus(ctx, manifestFileTableName, manifestId, f.UploadID, manifestFile.Uploaded, "")
		assert.NoError(t, err)
	}

	// Update status of manifest
	_, err = client.CheckUpdateManifestStatus(ctx, manifestFileTableName, manifestTableName, manifestId, m.Status)
	assert.NoError(t, err)

	// Check status should not have impacted manifest status
	m, err = client.GetManifestById(ctx, manifestTableName, manifestId)
	assert.NoError(t, err)
	assert.Equal(t, manifest.Initiated.String(), m.Status, "Check 3: Manifest status should be uploading")

	// SET TO IMPORTED --> Manifest status should be set to COMPLETED
	for _, f := range dtos {
		err := client.UpdateFileTableStatus(ctx, manifestFileTableName, manifestId, f.UploadID, manifestFile.Imported, "")
		assert.NoError(t, err)
	}

	// Update status of manifest
	_, err = client.CheckUpdateManifestStatus(ctx, manifestFileTableName, manifestTableName, manifestId, m.Status)
	assert.NoError(t, err)

	// Check status should not have impacted manifest status
	m, err = client.GetManifestById(ctx, manifestTableName, manifestId)
	assert.NoError(t, err)
	assert.Equal(t, manifest.Completed.String(), m.Status, "Check 3: Manifest status should be completed")

}

func testRemoveFailedFilesFromResponse(t *testing.T, _ *DynamoStore) {
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

func testGetAction(t *testing.T, _ *DynamoStore) {
	manifestId := "getActionTest"

	file := manifestFile.FileDTO{
		UploadID: "getActionTestId",
		Status:   manifestFile.Local,
	}

	// Check file that is newly uploaded and not in manifest Local (Unknown) --> Registered
	req, status, err := getWriteRequest(manifestId, file, manifestFile.Unknown)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Registered, status)
	assert.NotNil(t, req.PutRequest)

	// Check file that is removed locally and previously registered: Removed (Registered) --> Delete request
	file.Status = manifestFile.Removed
	req, status, err = getWriteRequest(manifestId, file, manifestFile.Registered)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Removed, status)
	assert.NotNil(t, req.DeleteRequest)

	// Failed (Failed) --> Registered
	file.Status = manifestFile.Failed
	req, status, err = getWriteRequest(manifestId, file, manifestFile.Registered)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Registered, status)
	assert.NotNil(t, req.PutRequest)

	// Imported (Finalized) --> Finalized
	file.Status = manifestFile.Imported
	req, status, err = getWriteRequest(manifestId, file, manifestFile.Finalized)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Verified, status)
	assert.NotNil(t, req.PutRequest)

	// Imported (Imported) --> Imported
	file.Status = manifestFile.Imported
	req, status, err = getWriteRequest(manifestId, file, manifestFile.Imported)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Imported, status)
	assert.Nil(t, req)

	// Registered (Registered) --> Registered
	file.Status = manifestFile.Registered
	req, status, err = getWriteRequest(manifestId, file, manifestFile.Registered)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Registered, status)
	assert.NotNil(t, req.PutRequest)

	// Registered (Finalized) --> Verified
	file.Status = manifestFile.Registered
	req, status, err = getWriteRequest(manifestId, file, manifestFile.Finalized)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Verified, status)
	assert.NotNil(t, req.PutRequest)

	// Registered (Imported) --> Verified
	file.Status = manifestFile.Registered
	req, status, err = getWriteRequest(manifestId, file, manifestFile.Imported)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Verified, status)
	assert.NotNil(t, req.PutRequest)

	// Registered (Verified) --> Verified
	file.Status = manifestFile.Registered
	req, status, err = getWriteRequest(manifestId, file, manifestFile.Verified)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Verified, status)
	assert.NotNil(t, req.PutRequest)

	// Finalized (Finalized) --> Finalized
	file.Status = manifestFile.Finalized
	req, status, err = getWriteRequest(manifestId, file, manifestFile.Finalized)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Finalized, status)
	assert.Nil(t, req)

	// Verified (Verified) --> Verified
	file.Status = manifestFile.Verified
	req, status, err = getWriteRequest(manifestId, file, manifestFile.Verified)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Verified, status)
	assert.Nil(t, req)

	// Unknown (Registered) --> Registered
	file.Status = manifestFile.Registered
	req, status, err = getWriteRequest(manifestId, file, manifestFile.Registered)
	assert.Nil(t, err, fmt.Sprintf("Could not get action for %v", file))
	assert.Equal(t, manifestFile.Registered, status)
	assert.NotNil(t, req.PutRequest)

}
