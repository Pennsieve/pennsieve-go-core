package dydb

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dydb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestManifestsStore(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, client *DynamoStore,
	){
		"get manifest by id":         testGetManifestById,
		"get manifests for dataset":  testGetManifestsForDataset,
		"update status for manifest": testUpdateManifestStatus,
	} {
		t.Run(scenario, func(t *testing.T) {
			client := getDynamoClient()
			store := NewDynamoStore(client)
			fn(t, store)
		})
	}
}

func testGetManifestById(t *testing.T, client *DynamoStore) {

	tb := dydb.ManifestTable{
		ManifestId:     "4444",
		DatasetId:      1,
		DatasetNodeId:  "N:Dataset:1234",
		OrganizationId: 1,
		UserId:         1,
		Status:         "Unknown",
		DateCreated:    time.Now().Unix(),
	}
	ctx := context.Background()

	err := client.CreateManifest(ctx, manifestTableName, tb)
	assert.Nil(t, err, "Manifest could not be created")

	out, err := client.GetManifestById(ctx, manifestTableName, tb.ManifestId)
	assert.Nil(t, err, "Manifest could not be fetched")
	assert.Equal(t, "N:Dataset:1234", out.DatasetNodeId)
}

func testGetManifestsForDataset(t *testing.T, client *DynamoStore) {

	ctx := context.Background()
	// Return multiple manifests
	datasetNodeId := "N:Dataset:5678"
	out, err := client.GetManifestsForDataset(ctx, manifestTableName, datasetNodeId)
	assert.Nil(t, err, "Manifest could not be fetched")
	assert.Equal(t, 2, len(out), "Incorrect number of manifests returned")

	// Return empty array for dataset without manifests
	nonExistingNodeId := "No:Dataset"
	out, err = client.GetManifestsForDataset(ctx, manifestTableName, nonExistingNodeId)
	assert.Nil(t, err, "Manifest could not be fetched")
	assert.Equal(t, 0, len(out), "Incorrect number of manifests returned")

}

func testUpdateManifestStatus(t *testing.T, client *DynamoStore) {
	ctx := context.Background()
	manifestId := "1111"

	err := client.updateManifestStatus(ctx, manifestTableName, manifestId, manifest.Completed)
	assert.Nil(t, err, "Manifest status could not be updated")

	out, err := client.GetManifestById(ctx, manifestTableName, manifestId)
	assert.Nil(t, err, "Manifest could not be fetched")
	assert.Equal(t, "Completed", out.Status)
}
