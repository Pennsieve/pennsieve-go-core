package pgdb

import (
	"context"
	"database/sql"
	"github.com/pennsieve/pennsieve-go-core/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDatasetContributor(t *testing.T) {
	orgId := 3
	db := testDB[orgId]
	store := NewSQLStore(db)

	addTestDataset(db, "Test Dataset - AddDatasetContributor")
	defer test.Truncate(t, db, orgId, "datasets")

	//addTestDataset(db, "Test Dataset - with Contributors")
	//defer test.Truncate(t, db, orgId, "datasets")

	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Add Dataset Contributor": testAddDatasetContributor,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := orgId
			store := store
			fn(t, store, orgId)
		})
	}
}

func testAddDatasetContributor(t *testing.T, store *SQLStore, orgId int) {
	// get the dataset
	datasetName := "Test Dataset - AddDatasetContributor"
	ds, err := store.GetDatasetByName(context.TODO(), datasetName)
	assert.NoError(t, err)
	assert.Equal(t, datasetName, ds.Name)

	// get the contributor
	userId := int64(1004)
	contributor, err := store.GetContributorByUserId(context.TODO(), userId)
	assert.NoError(t, err)
	assert.Equal(t, sql.NullInt64{Int64: userId, Valid: true}, contributor.UserId)

	// add the dataset contributor
	datasetContributor, err := store.AddDatasetContributor(context.TODO(), ds, contributor)
	assert.NoError(t, err)
	assert.Equal(t, ds.Id, datasetContributor.DatasetId)
	assert.Equal(t, contributor.Id, datasetContributor.ContributorId)
}
