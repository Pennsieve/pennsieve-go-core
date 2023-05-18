package pgdb

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrganizations(t *testing.T) {
	orgId := 0
	store := NewSQLStore(testDB[orgId])

	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Get Organization by Id":     testGetOrganizationById,
		"Get Organization by NodeId": testGetOrganizationByNodeId,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := orgId
			store := store
			fn(t, store, orgId)
		})
	}
}

func testGetOrganizationById(t *testing.T, store *SQLStore, orgId int) {
	id := int64(42)

	org, err := store.GetOrganization(context.TODO(), id)
	assert.NoError(t, err)
	assert.Equal(t, id, org.Id)
}

func testGetOrganizationByNodeId(t *testing.T, store *SQLStore, orgId int) {
	nodeId := "N:organization:2b809c6f-9941-47a2-9593-9540fbe77ff1"

	org, err := store.GetOrganizationByNodeId(context.TODO(), nodeId)
	assert.NoError(t, err)
	assert.Equal(t, nodeId, org.NodeId)
}
