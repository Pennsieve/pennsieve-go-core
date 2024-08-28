package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/datasetType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDatasetRelease(t *testing.T) {
	orgId := 3
	db := testDB[orgId]
	store := NewSQLStore(db)

	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Add Dataset Release":       testAddDatasetRelease,
		"Get Dataset Release":       testGetDatasetRelease,
		"Get Dataset Release by Id": testGetDatasetReleaseById,
		"Update Dataset Release":    testUpdateDatasetRelease,
		"Update Release Status":     testUpdateReleaseStatus,
		"Update Publishing Status":  testUpdatePublishingStatus,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := orgId
			store := store
			fn(t, store, orgId)
		})
	}
}

func deleteDataset(store *SQLStore, datasetId int64) {
	statement := fmt.Sprintf("DELETE FROM datasets WHERE id = $1")
	_, err := store.db.ExecContext(context.TODO(),
		statement,
		datasetId,
	)

	if err != nil {
		fmt.Printf(fmt.Sprintf("deleteDataset() database error: %v", err))
	}
}

func addDatasetTypeRelease(store *SQLStore, orgId int, name string) int64 {
	var err error
	defaultDatasetStatus, err := store.GetDefaultDatasetStatus(context.TODO(), orgId)
	if err != nil {
		panic("testCreateDataset(): failed to get default dataset status")
	}
	defaultDataUseAgreement, err := store.GetDefaultDataUseAgreement(context.TODO(), orgId)
	if err != nil {
		panic("testCreateDataset(): failed to get default data use agreement")
	}
	createDatasetParams := CreateDatasetParams{
		Name:                         name,
		Description:                  name,
		Status:                       defaultDatasetStatus,
		AutomaticallyProcessPackages: false,
		License:                      "Community Data License Agreement – Sharing",
		Tags:                         nil,
		DataUseAgreement:             defaultDataUseAgreement,
		Type:                         datasetType.Release,
	}

	dataset, err := store.CreateDataset(context.TODO(), createDatasetParams)
	if err != nil {
		panic(err)
	}
	return dataset.Id
}

func testAddDatasetRelease(t *testing.T, store *SQLStore, orgId int) {
	datasetId := addDatasetTypeRelease(store, orgId, "Test Dataset Release - 01")

	input := pgdb.DatasetRelease{
		DatasetId:        datasetId,
		Origin:           "GitHub",
		Url:              "https://github.com/pennsieve/pennsieve",
		Label:            sql.NullString{},
		Marker:           sql.NullString{},
		Properties:       nil,
		Tags:             nil,
		ReleaseDate:      sql.NullTime{},
		ReleaseStatus:    "created",
		PublishingStatus: "initial",
	}

	output, err := store.AddDatasetRelease(context.TODO(), input)
	assert.NoError(t, err)
	assert.Equal(t, datasetId, output.DatasetId)
	deleteDataset(store, datasetId)
}

func testGetDatasetRelease(t *testing.T, store *SQLStore, orgId int) {
	datasetId := addDatasetTypeRelease(store, orgId, "Test Dataset Release - 02")

	label := "v1.0.0"
	marker := "0123456"

	input := pgdb.DatasetRelease{
		DatasetId:        datasetId,
		Origin:           "GitHub",
		Url:              "https://github.com/pennsieve/pennsieve",
		Label:            sql.NullString{Valid: true, String: label},
		Marker:           sql.NullString{Valid: true, String: marker},
		Properties:       nil,
		Tags:             nil,
		ReleaseDate:      sql.NullTime{},
		ReleaseStatus:    "created",
		PublishingStatus: "initial",
	}

	output, err := store.AddDatasetRelease(context.TODO(), input)
	assert.NoError(t, err)
	assert.Equal(t, datasetId, output.DatasetId)

	release, err := store.GetDatasetRelease(context.TODO(), output.DatasetId, label, marker)
	assert.NoError(t, err)
	assert.True(t, release.Label.Valid)
	assert.Equal(t, label, release.Label.String)
	assert.True(t, release.Marker.Valid)
	assert.Equal(t, marker, release.Marker.String)
	deleteDataset(store, datasetId)
}

func testGetDatasetReleaseById(t *testing.T, store *SQLStore, orgId int) {
	datasetId := addDatasetTypeRelease(store, orgId, "Test Dataset Release - 03")

	label := "v2.0.0"
	marker := "1234567"

	input := pgdb.DatasetRelease{
		DatasetId:        datasetId,
		Origin:           "GitHub",
		Url:              "https://github.com/pennsieve/pennsieve",
		Label:            sql.NullString{Valid: true, String: label},
		Marker:           sql.NullString{Valid: true, String: marker},
		Properties:       nil,
		Tags:             nil,
		ReleaseDate:      sql.NullTime{},
		ReleaseStatus:    "created",
		PublishingStatus: "initial",
	}

	output, err := store.AddDatasetRelease(context.TODO(), input)
	assert.NoError(t, err)
	assert.Equal(t, datasetId, output.DatasetId)
	assert.True(t, output.Id > 0)

	release, err := store.GetDatasetReleaseById(context.TODO(), output.Id)
	assert.NoError(t, err)
	assert.True(t, release.Label.Valid)
	assert.Equal(t, label, release.Label.String)
	assert.True(t, release.Marker.Valid)
	assert.Equal(t, marker, release.Marker.String)
	deleteDataset(store, datasetId)
}

func testUpdateDatasetRelease(t *testing.T, store *SQLStore, orgId int) {
	datasetId := addDatasetTypeRelease(store, orgId, "Test Dataset Release - 04")

	input := pgdb.DatasetRelease{
		DatasetId:        datasetId,
		Origin:           "GitHub",
		Url:              "https://github.com/pennsieve/pennsieve",
		Label:            sql.NullString{},
		Marker:           sql.NullString{},
		Properties:       nil,
		Tags:             nil,
		ReleaseDate:      sql.NullTime{},
		ReleaseStatus:    "created",
		PublishingStatus: "initial",
	}

	output, err := store.AddDatasetRelease(context.TODO(), input)
	assert.NoError(t, err)
	assert.Equal(t, datasetId, output.DatasetId)

	label := "v4.0.0"
	marker := "1010101"
	releaseDate := time.Date(2024, time.August, 27, 18, 32, 25, 0, time.UTC)

	update := pgdb.DatasetRelease{
		Id:               output.Id,
		DatasetId:        datasetId,
		Origin:           "GitHub",
		Url:              "https://github.com/pennsieve/pennsieve",
		Label:            sql.NullString{Valid: true, String: label},
		Marker:           sql.NullString{Valid: true, String: marker},
		Properties:       nil,
		Tags:             nil,
		ReleaseDate:      sql.NullTime{Valid: true, Time: releaseDate},
		ReleaseStatus:    "created",
		PublishingStatus: "initial",
	}

	updated, err := store.UpdateDatasetRelease(context.TODO(), update)
	assert.NoError(t, err)
	assert.Equal(t, datasetId, updated.DatasetId)
	assert.Equal(t, output.Id, updated.Id)
	assert.Equal(t, label, updated.Label.String)
	assert.Equal(t, marker, updated.Marker.String)
	assert.True(t, releaseDate.Equal(updated.ReleaseDate.Time))
	deleteDataset(store, datasetId)
}

func testUpdateReleaseStatus(t *testing.T, store *SQLStore, orgId int) {
	datasetId := addDatasetTypeRelease(store, orgId, "Test Dataset Release - 05")

	initialReleaseStatus := "prerelease"
	input := pgdb.DatasetRelease{
		DatasetId:        datasetId,
		Origin:           "GitHub",
		Url:              "https://github.com/pennsieve/pennsieve",
		Label:            sql.NullString{},
		Marker:           sql.NullString{},
		Properties:       nil,
		Tags:             nil,
		ReleaseDate:      sql.NullTime{},
		ReleaseStatus:    initialReleaseStatus,
		PublishingStatus: "initial",
	}

	output, err := store.AddDatasetRelease(context.TODO(), input)
	assert.NoError(t, err)
	assert.Equal(t, datasetId, output.DatasetId)

	label := "v4.0.0"
	marker := "1010101"
	releaseDate := time.Date(2024, time.August, 27, 18, 32, 25, 0, time.UTC)

	updatedReleaseStatus := "published"
	update := pgdb.DatasetRelease{
		Id:               output.Id,
		DatasetId:        datasetId,
		Origin:           "GitHub",
		Url:              "https://github.com/pennsieve/pennsieve",
		Label:            sql.NullString{Valid: true, String: label},
		Marker:           sql.NullString{Valid: true, String: marker},
		Properties:       nil,
		Tags:             nil,
		ReleaseDate:      sql.NullTime{Valid: true, Time: releaseDate},
		ReleaseStatus:    updatedReleaseStatus,
		PublishingStatus: "initial",
	}

	updated, err := store.UpdateDatasetRelease(context.TODO(), update)
	assert.NoError(t, err)
	assert.Equal(t, datasetId, updated.DatasetId)
	assert.Equal(t, output.Id, updated.Id)
	assert.Equal(t, label, updated.Label.String)
	assert.Equal(t, marker, updated.Marker.String)
	assert.True(t, releaseDate.Equal(updated.ReleaseDate.Time))
	assert.Equal(t, updatedReleaseStatus, updated.ReleaseStatus)
	deleteDataset(store, datasetId)
}

func testUpdatePublishingStatus(t *testing.T, store *SQLStore, orgId int) {
	datasetId := addDatasetTypeRelease(store, orgId, "Test Dataset Release - 06")

	initialReleaseStatus := "created"
	initialPublishingStatus := "initial"
	input := pgdb.DatasetRelease{
		DatasetId:        datasetId,
		Origin:           "GitHub",
		Url:              "https://github.com/pennsieve/pennsieve",
		Label:            sql.NullString{},
		Marker:           sql.NullString{},
		Properties:       nil,
		Tags:             nil,
		ReleaseDate:      sql.NullTime{},
		ReleaseStatus:    initialReleaseStatus,
		PublishingStatus: initialPublishingStatus,
	}

	output, err := store.AddDatasetRelease(context.TODO(), input)
	assert.NoError(t, err)
	assert.Equal(t, datasetId, output.DatasetId)

	label := "v4.0.0"
	marker := "1010101"
	releaseDate := time.Date(2024, time.August, 27, 18, 32, 25, 0, time.UTC)

	updatedReleaseStatus := "created"
	updatedPublishingStatus := "succeeded"
	update := pgdb.DatasetRelease{
		Id:               output.Id,
		DatasetId:        datasetId,
		Origin:           "GitHub",
		Url:              "https://github.com/pennsieve/pennsieve",
		Label:            sql.NullString{Valid: true, String: label},
		Marker:           sql.NullString{Valid: true, String: marker},
		Properties:       nil,
		Tags:             nil,
		ReleaseDate:      sql.NullTime{Valid: true, Time: releaseDate},
		ReleaseStatus:    updatedReleaseStatus,
		PublishingStatus: updatedPublishingStatus,
	}

	updated, err := store.UpdateDatasetRelease(context.TODO(), update)
	assert.NoError(t, err)
	assert.Equal(t, datasetId, updated.DatasetId)
	assert.Equal(t, output.Id, updated.Id)
	assert.Equal(t, label, updated.Label.String)
	assert.Equal(t, marker, updated.Marker.String)
	assert.True(t, releaseDate.Equal(updated.ReleaseDate.Time))
	assert.Equal(t, updatedReleaseStatus, updated.ReleaseStatus)
	assert.Equal(t, updatedPublishingStatus, updated.PublishingStatus)
	deleteDataset(store, datasetId)
}
