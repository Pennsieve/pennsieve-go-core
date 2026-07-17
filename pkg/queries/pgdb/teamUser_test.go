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
		"Get Team Memberships":         testGetTeamMemberships,
		"Get Team Claims":              testGetTeamClaims,
		"Get Team Memberships For Org": testGetTeamMembershipsForOrg,
		"Get Team Claims For Org":      testGetTeamClaimsForOrg,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := 1
			store := NewSQLStore(testDB[orgId])
			fn(t, store, orgId)
		})
	}
}

func testGetTeamMemberships(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1001)

	memberships, err := store.GetTeamMemberships(context.TODO(), userId)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(memberships))
}

func testGetTeamClaims(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1001)

	claims, err := store.GetTeamClaims(context.TODO(), userId)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(claims))
}

func testGetTeamMembershipsForOrg(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1001)

	// The seeded user's teams live in org 1 (their preferred org) -> same result as the
	// deprecated method when the correct org is passed.
	memberships, err := store.GetTeamMembershipsForOrg(context.TODO(), userId, int64(orgId))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(memberships))

	// A different org id drives the filter independently of preferred_org_id -> no results.
	memberships, err = store.GetTeamMembershipsForOrg(context.TODO(), userId, int64(9999))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(memberships))
}

func testGetTeamClaimsForOrg(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1001)

	claims, err := store.GetTeamClaimsForOrg(context.TODO(), userId, int64(orgId))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(claims))

	claims, err = store.GetTeamClaimsForOrg(context.TODO(), userId, int64(9999))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(claims))
}
