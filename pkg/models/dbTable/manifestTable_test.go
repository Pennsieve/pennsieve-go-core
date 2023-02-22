package dbTable

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestManifest(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, client *dynamodb.Client,
	){
		"get manifest by id":         testGetManifestById,
		"get manifests for dataset":  testGetManifestsForDataset,
		"update status for manifest": testUpdateManifestStatus,
	} {
		t.Run(scenario, func(t *testing.T) {
			client := getDynamoClient()
			fn(t, client)
		})
	}
}

func testGetManifestById(t *testing.T, client *dynamodb.Client) {

	tb := ManifestTable{
		ManifestId:     "4444",
		DatasetId:      1,
		DatasetNodeId:  "N:Dataset:1234",
		OrganizationId: 1,
		UserId:         1,
		Status:         "Unknown",
		DateCreated:    time.Now().Unix(),
	}

	err := tb.CreateManifest(client, manifestTableName, tb)
	assert.Nil(t, err, "Manifest could not be created")

	out, err := tb.GetFromManifest(client, manifestTableName, tb.ManifestId)
	assert.Nil(t, err, "Manifest could not be fetched")
	assert.Equal(t, "N:Dataset:1234", out.DatasetNodeId)
}

func testGetManifestsForDataset(t *testing.T, client *dynamodb.Client) {

	var m *ManifestTable

	// Return multiple manifests
	datasetNodeId := "N:Dataset:5678"
	out, err := m.GetManifestsForDataset(client, manifestTableName, datasetNodeId)
	assert.Nil(t, err, "Manifest could not be fetched")
	assert.Equal(t, 2, len(out), "Incorrect number of manifests returned")

	// Return empty array for dataset without manifests
	nonExisitingNodeId := "No:Dataset"
	out, err = m.GetManifestsForDataset(client, manifestTableName, nonExisitingNodeId)
	assert.Nil(t, err, "Manifest could not be fetched")
	assert.Equal(t, 0, len(out), "Incorrect number of manifests returned")

}

func testUpdateManifestStatus(t *testing.T, client *dynamodb.Client) {
	var m *ManifestTable
	manifestId := "1111"

	err := m.UpdateManifestStatus(client, manifestTableName, manifestId, manifest.Completed)
	assert.Nil(t, err, "Manifest status could not be updated")

	out, err := m.GetFromManifest(client, manifestTableName, manifestId)
	assert.Nil(t, err, "Manifest could not be fetched")
	assert.Equal(t, "Completed", out.Status)
}
