package pgdb

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrganizationUser(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Get Organization User":           testGetOrganizationUser,
		"Add Organization User":           testAddOrganizationUser,
		"Add Existing OrgUser Membership": testAddExistingOrgUserMembership,
		"Add Guest to Organization":       testAddGuestToOrganization,
		"Get Organization Claim":          testGetOrganizationClaim,
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

func testAddOrganizationUser(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1001)
	organizationId := int64(2)
	permissionBit := pgdb.DbPermission(pgdb.Delete)

	orgUser, err := store.AddOrganizationUser(context.TODO(), organizationId, userId, permissionBit)
	assert.NoError(t, err)
	assert.Equal(t, userId, orgUser.UserId)
	assert.Equal(t, organizationId, orgUser.OrganizationId)
	assert.Equal(t, permissionBit, orgUser.DbPermission)
}

// testAddExistingOrgUserMembership should not update the user's organization membership
func testAddExistingOrgUserMembership(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1001)
	organizationId := int64(1)
	permissionBit := pgdb.DbPermission(pgdb.Administer)

	orgUser, err := store.AddOrganizationUser(context.TODO(), organizationId, userId, permissionBit)
	assert.NoError(t, err)
	assert.Equal(t, userId, orgUser.UserId)
	assert.Equal(t, organizationId, orgUser.OrganizationId)
	assert.NotEqual(t, permissionBit, orgUser.DbPermission)
}

func testAddGuestToOrganization(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1001)
	organizationId := int64(3)
	permissionBit := pgdb.DbPermission(pgdb.Guest)

	orgUser, err := store.AddOrganizationUser(context.TODO(), organizationId, userId, permissionBit)
	assert.NoError(t, err)
	assert.Equal(t, userId, orgUser.UserId)
	assert.Equal(t, organizationId, orgUser.OrganizationId)
	assert.Equal(t, permissionBit, orgUser.DbPermission)
}

func testGetOrganizationClaim(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1001)
	organizationId := int64(1)

	// get the organization so that we can check NodeId
	org, err := store.GetOrganization(context.TODO(), organizationId)
	assert.NoError(t, err)
	assert.Equal(t, organizationId, org.Id)

	// get the organization claim
	orgClaim, err := store.GetOrganizationClaim(context.TODO(), userId, organizationId)
	assert.NoError(t, err)
	assert.Equal(t, organizationId, orgClaim.IntId)
	assert.Equal(t, org.NodeId, orgClaim.NodeId)
}
