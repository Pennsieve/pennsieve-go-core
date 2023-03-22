package pgdb

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDatasetStorage(t *testing.T) {

	orgId := 1
	store := NewSQLStore(testDB[orgId])
	datasetId := int64(1)
	defer test.Truncate(t, store.db, orgId, "dataset_storage")

	// Adding 10
	err := store.IncrementDatasetStorage(context.Background(), datasetId, 10)
	assert.NoError(t, err)

	actualSize, err := store.GetDatasetStorageById(context.Background(), datasetId)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), actualSize, "Size is expected to be 10")

	// Adding 10
	err = store.IncrementDatasetStorage(context.Background(), datasetId, 10)
	assert.NoError(t, err)

	actualSize, err = store.GetDatasetStorageById(context.Background(), datasetId)
	assert.NoError(t, err)
	assert.Equal(t, int64(20), actualSize, "Size is expected to be 20")

	// Removing 10
	err = store.IncrementDatasetStorage(context.Background(), datasetId, -10)
	assert.NoError(t, err)

	actualSize, err = store.GetDatasetStorageById(context.Background(), datasetId)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), actualSize, "Size is expected to be 10")

}
