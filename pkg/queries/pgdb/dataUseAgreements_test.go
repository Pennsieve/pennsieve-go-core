package pgdb

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataUseAgreements(t *testing.T) {
	orgId := 2
	db := testDB[orgId]
	store := NewSQLStore(db)

	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Get Default Data Use Agreement": testGetDefaultDataUseAgreement,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := orgId
			store := store
			fn(t, store, orgId)
		})
	}
}

func testGetDefaultDataUseAgreement(t *testing.T, store *SQLStore, orgId int) {
	expectedId := int64(1001)
	dataUseAgreement, err := store.GetDefaultDataUseAgreement(context.TODO(), orgId)
	assert.NoError(t, err)
	assert.Equal(t, expectedId, dataUseAgreement.Id)
}
