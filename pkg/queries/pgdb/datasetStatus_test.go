package pgdb

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDatasetStatus(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Get Default Dataset Status": testGetDefaultDatasetStatus,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := 1
			store := NewSQLStore(testDB[orgId])
			fn(t, store, orgId)
		})
	}
}

func testGetDefaultDatasetStatus(t *testing.T, store *SQLStore, orgId int) {
	expectedId := int64(1)
	datasetStatus, err := store.GetDefaultDatasetStatus(context.TODO(), orgId)
	assert.NoError(t, err)
	assert.Equal(t, expectedId, datasetStatus.Id)
}
