package pgdb

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"slices"
	"testing"
)

func TestOrganizationUser(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Get Organization User":                                           testGetOrganizationUser,
		"Add Organization User":                                           testAddOrganizationUser,
		"Add Existing OrgUser Membership":                                 testAddExistingOrgUserMembership,
		"Add Guest to Organization":                                       testAddGuestToOrganization,
		"Get Organization Claim":                                          testGetOrganizationClaim,
		"Get Organization Claim error when no org user":                   testGetOrganizationClaimNoOrgUser,
		"Get Organization Claim when org has no feature flags":            testGetOrganizationClaimNoFeatureFlags,
		"Get Organization Claim when org has many feature flags":          testGetOrganizationClaimManyFeatureFlags,
		"Get Organization Claim when org has only disabled feature flags": testGetOrganizationClaimAllDisabled,
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

	// get the organization claim by org id
	orgClaim, err := store.GetOrganizationClaim(context.TODO(), userId, organizationId)
	require.NoError(t, err)
	assert.Equal(t, organizationId, orgClaim.IntId)
	assert.Equal(t, org1NodeId, orgClaim.NodeId)
	assert.Equal(t, pgdb.Delete, orgClaim.Role)
	// org 1 is the sandbox org created by migrations in the seed DB.
	// It has one feature flag
	assert.Len(t, orgClaim.EnabledFeatures, 1)
	assert.Equal(t, "sandbox_org_feature", orgClaim.EnabledFeatures[0].Feature)
	assert.Equal(t, organizationId, orgClaim.EnabledFeatures[0].OrganizationId)
	assert.True(t, orgClaim.EnabledFeatures[0].Enabled)

	// Get org claim by node id
	{
		orgClaim, err := store.GetOrganizationClaimByNodeId(context.TODO(), userId, org1NodeId)
		require.NoError(t, err)
		assert.Equal(t, organizationId, orgClaim.IntId)
		assert.Equal(t, org1NodeId, orgClaim.NodeId)
		assert.Equal(t, pgdb.Delete, orgClaim.Role)
		// org 1 is the sandbox org created by migrations in the seed DB.
		// It has one feature flag
		assert.Len(t, orgClaim.EnabledFeatures, 1)
		assert.Equal(t, "sandbox_org_feature", orgClaim.EnabledFeatures[0].Feature)
		assert.Equal(t, organizationId, orgClaim.EnabledFeatures[0].OrganizationId)
		assert.True(t, orgClaim.EnabledFeatures[0].Enabled)
	}
}

func testGetOrganizationClaimNoOrgUser(t *testing.T, store *SQLStore, _ int) {
	// store_test.go and tests in this file do not add user 1002 to org 1
	userId := int64(1002)
	organizationId := int64(1)

	_, err := store.GetOrganizationClaim(context.TODO(), userId, organizationId)
	assert.ErrorContains(t, err, "organization user was not found")

	_, byNodeIdErr := store.GetOrganizationClaimByNodeId(context.TODO(), userId, org1NodeId)
	assert.ErrorContains(t, byNodeIdErr, "organization user was not found")
}

func testGetOrganizationClaimNoFeatureFlags(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1002)
	organizationId := int64(2)

	// get the organization claim
	orgClaim, err := store.GetOrganizationClaim(context.TODO(), userId, organizationId)
	require.NoError(t, err)
	assert.Equal(t, organizationId, orgClaim.IntId)
	assert.Equal(t, org2NodeId, orgClaim.NodeId)
	assert.Equal(t, pgdb.Delete, orgClaim.Role)
	assert.Empty(t, orgClaim.EnabledFeatures)

	byNodeIdClaim, err := store.GetOrganizationClaimByNodeId(context.TODO(), userId, org2NodeId)
	require.NoError(t, err)
	assert.Equal(t, orgClaim, byNodeIdClaim)

}

func testGetOrganizationClaimManyFeatureFlags(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(3402)
	organizationId := int64(402)

	// get the organization claim by org id
	orgClaim, err := store.GetOrganizationClaim(context.TODO(), userId, organizationId)
	require.NoError(t, err)
	assert.Equal(t, organizationId, orgClaim.IntId)
	assert.Equal(t, org402NodeId, orgClaim.NodeId)
	assert.Equal(t, pgdb.Read, orgClaim.Role)
	assert.Len(t, orgClaim.EnabledFeatures, 4)
	enabledFeatures := []string{"one", "two", "three", "four"}
	for _, enabledFeature := range enabledFeatures {
		index := slices.IndexFunc(orgClaim.EnabledFeatures, func(flag pgdb.FeatureFlags) bool {
			return flag.Enabled && flag.Feature == fmt.Sprintf("feature %s", enabledFeature)
		})
		assert.True(t, index >= 0, "expected enabled feature %s not found in %s", enabledFeature, orgClaim.EnabledFeatures)

	}

	// get org claim by node id
	{
		orgClaim, err := store.GetOrganizationClaimByNodeId(context.TODO(), userId, org402NodeId)
		require.NoError(t, err)
		assert.Equal(t, organizationId, orgClaim.IntId)
		assert.Equal(t, org402NodeId, orgClaim.NodeId)
		assert.Equal(t, pgdb.Read, orgClaim.Role)
		assert.Len(t, orgClaim.EnabledFeatures, 4)
		enabledFeatures := []string{"one", "two", "three", "four"}
		for _, enabledFeature := range enabledFeatures {
			index := slices.IndexFunc(orgClaim.EnabledFeatures, func(flag pgdb.FeatureFlags) bool {
				return flag.Enabled && flag.Feature == fmt.Sprintf("feature %s", enabledFeature)
			})
			assert.True(t, index >= 0, "expected enabled feature %s not found in %s", enabledFeature, orgClaim.EnabledFeatures)

		}
	}
}

func testGetOrganizationClaimAllDisabled(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(3403)
	organizationId := int64(403)

	// get the organization claim by org id
	orgClaim, err := store.GetOrganizationClaim(context.TODO(), userId, organizationId)
	require.NoError(t, err)
	assert.Equal(t, organizationId, orgClaim.IntId)
	assert.Equal(t, org403NodeId, orgClaim.NodeId)
	assert.Equal(t, pgdb.Read, orgClaim.Role)
	assert.Empty(t, orgClaim.EnabledFeatures)

	// get org claim by node id
	{
		orgClaim, err := store.GetOrganizationClaimByNodeId(context.TODO(), userId, org403NodeId)
		require.NoError(t, err)
		assert.Equal(t, organizationId, orgClaim.IntId)
		assert.Equal(t, org403NodeId, orgClaim.NodeId)
		assert.Equal(t, pgdb.Read, orgClaim.Role)
		assert.Empty(t, orgClaim.EnabledFeatures)
	}
}
