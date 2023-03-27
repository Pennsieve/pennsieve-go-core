package pgdb

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrganizationUser(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Get Organization User": testGetOrganizationUser,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := 1
			store := NewSQLStore(testDB[orgId])
			fn(t, store, orgId)
		})
	}
}

func testGetOrganizationUser(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1001)
	organizationId := int64(1)

	orgUser, err := store.GetOrganizationUser(context.TODO(), organizationId, userId)
	assert.NoError(t, err)
	assert.Equal(t, userId, orgUser.UserId)
	assert.Equal(t, organizationId, orgUser.OrganizationId)
}
