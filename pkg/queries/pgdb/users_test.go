package pgdb

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUsers is the main Test Suite function for Users.
func TestUsers(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Get User by Id":         testGetUserByID,
		"Get User by Cognito Id": testGetUserByCognitoId,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := 0
			store := NewSQLStore(testDB[orgId])
			fn(t, store, orgId)
		})
	}
}

func testGetUserByID(t *testing.T, store *SQLStore, orgId int) {
	id := int64(1001)
	user, err := store.GetUserById(context.TODO(), id)
	assert.NoError(t, err)
	assert.Equal(t, user.Id, id)
}

func testGetUserByCognitoId(t *testing.T, store *SQLStore, orgId int) {
	id := int64(1002)
	cognitoId := "22222222-2222-2222-2222-222222222222"
	user, err := store.GetByCognitoId(context.TODO(), cognitoId)
	assert.NoError(t, err)
	assert.Equal(t, user.Id, id)
}
