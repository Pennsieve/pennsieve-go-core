package pgdb

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTokens(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Get Token User by Cognito Id": testGetTokenUserByCognitoId,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := 0
			store := NewSQLStore(testDB[orgId])
			fn(t, store, orgId)
		})
	}
}

func testGetTokenUserByCognitoId(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(2001)
	userNodeId := "N:user:2001"
	organizationId := int64(1)
	cognitoId := "00000000-1111-0000-2222-000000002001"
	user, err := store.GetUserByCognitoId(context.TODO(), cognitoId)
	assert.NoError(t, err)
	assert.Equal(t, userId, user.Id)
	assert.Equal(t, userNodeId, user.NodeId)
	assert.Equal(t, organizationId, user.PreferredOrg)
}
