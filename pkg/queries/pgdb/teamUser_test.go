package pgdb

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
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

	// Index by team id so value assertions don't depend on row order (the query has no
	// ORDER BY). Verifying the scanned fields guards against a column/field ordering mismatch
	// in the Scan, which a length-only check would miss.
	byTeam := make(map[int64]UserTeamMembership, len(memberships))
	for _, m := range memberships {
		// Columns common to every row (user + org side of the join).
		assert.Equal(t, userId, m.UserId)
		assert.Equal(t, "user1@pennsieve.org", m.UserEmail)
		assert.Equal(t, "N:user:1", m.UserNodeId)
		assert.Equal(t, int64(orgId), m.OrgId)
		assert.Equal(t, pgdb.Delete, m.OrgUserPermission)
		byTeam[m.TeamId] = m
	}

	if research, ok := byTeam[researchTeamId]; assert.True(t, ok, "expected research team membership") {
		assert.Equal(t, researchTeamName, research.TeamName)
		assert.Equal(t, researchTeamNodeId, research.TeamNodeId)
		assert.Equal(t, pgdb.Administer, research.TeamPermission)
		assert.False(t, research.TeamType.Valid) // seeded with an empty system_team_type -> NULL
	}

	if publishers, ok := byTeam[publishingTeamId]; assert.True(t, ok, "expected publishing team membership") {
		assert.Equal(t, publishingTeamName, publishers.TeamName)
		assert.Equal(t, publishingTeamNodeId, publishers.TeamNodeId)
		assert.Equal(t, pgdb.Administer, publishers.TeamPermission)
		assert.True(t, publishers.TeamType.Valid)
		assert.Equal(t, "publishers", publishers.TeamType.String)
	}

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

	// Verify the membership -> Claim mapping (including the null system_team_type -> "<none>").
	byTeam := make(map[int64]teamUser.Claim, len(claims))
	for _, c := range claims {
		byTeam[c.IntId] = c
	}

	if research, ok := byTeam[researchTeamId]; assert.True(t, ok, "expected research team claim") {
		assert.Equal(t, researchTeamName, research.Name)
		assert.Equal(t, researchTeamNodeId, research.NodeId)
		assert.Equal(t, pgdb.Administer, research.Permission)
		assert.Equal(t, "<none>", research.TeamType)
	}

	if publishers, ok := byTeam[publishingTeamId]; assert.True(t, ok, "expected publishing team claim") {
		assert.Equal(t, publishingTeamName, publishers.Name)
		assert.Equal(t, publishingTeamNodeId, publishers.NodeId)
		assert.Equal(t, pgdb.Administer, publishers.Permission)
		assert.Equal(t, "publishers", publishers.TeamType)
	}

	claims, err = store.GetTeamClaimsForOrg(context.TODO(), userId, int64(9999))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(claims))
}
