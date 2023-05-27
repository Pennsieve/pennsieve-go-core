package pgdb

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTeamUser(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Get Publishers Claim": testGetPublishersClaim,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := 1
			store := NewSQLStore(testDB[orgId])
			fn(t, store, orgId)
		})
	}
}

func testGetPublishersClaim(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1001)
	organizationId := int64(1)

	claim, err := store.GetPublishersClaim(context.TODO(), organizationId, userId)
	assert.NoError(t, err)
	assert.Equal(t, publishingTeamId, claim.IntId)
	assert.Equal(t, publishingTeamName, claim.Name)
	assert.Equal(t, publishingTeamNodeId, claim.NodeId)
}
