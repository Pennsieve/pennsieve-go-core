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
		"Get Team Memberships": testGetTeamMemberships,
		"Get Team Claims":      testGetTeamClaims,
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
