package pgdb

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContributors(t *testing.T) {
	orgId := 3
	db := testDB[orgId]
	store := NewSQLStore(db)

	//addTestDataset(db, "Test Dataset - with Contributors")
	//defer test.Truncate(t, db, orgId, "datasets")

	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Add Contributor Existing User": testAddContributorExistingUser,
		"Add Contributor External":      testAddContributorExternal,
		"Get Contributor By Id":         testGetContributorById,
		"Get Contributor By User Id":    testGetContributorByUserId,
		"Get Contributor By Email":      testGetContributorByEmail,
		"Get Contributor By Orcid":      testGetContributorByOrcid,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := orgId
			store := store
			fn(t, store, orgId)
		})
	}
}

func testAddContributorExistingUser(t *testing.T, store *SQLStore, orgId int) {
	// get an existing user
	user, err := store.GetUserById(context.TODO(), 1003)
	assert.NoError(t, err)

	newContributor := NewContributor{
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		EmailAddress: user.Email,
		UserId:       user.Id,
	}

	contributor, err := store.AddContributor(context.TODO(), newContributor)
	assert.NoError(t, err)
	assert.Equal(t, user.FirstName, contributor.FirstName)
	assert.Equal(t, user.LastName, contributor.LastName)
	assert.Equal(t, user.Email, contributor.Email)
	assert.Equal(t, sql.NullInt64{Int64: user.Id, Valid: true}, contributor.UserId)
}

func testAddContributorExternal(t *testing.T, store *SQLStore, orgId int) {
	newContributor := NewContributor{
		FirstName:     "Someone",
		MiddleInitial: "X",
		LastName:      "Ternal",
		Degree:        "MD",
		EmailAddress:  "someone.x.ternal@not-pennsieve.org",
		Orcid:         "0000-0000-0000-4321",
	}

	contributor, err := store.AddContributor(context.TODO(), newContributor)
	assert.NoError(t, err)
	assert.Equal(t, newContributor.EmailAddress, contributor.Email)
	assert.Equal(t, sql.NullString{String: newContributor.Orcid, Valid: true}, contributor.Orcid)
}

func testGetContributorById(t *testing.T, store *SQLStore, orgId int) {
	contributorId := int64(1)
	contributor, err := store.GetContributor(context.TODO(), contributorId)
	assert.NoError(t, err)
	assert.Equal(t, contributorId, contributor.Id)
	assert.Equal(t, "user4@pennsieve.org", contributor.Email)
}

func testGetContributorByUserId(t *testing.T, store *SQLStore, orgId int) {
	userId := int64(1004)
	contributor, err := store.GetContributorByUserId(context.TODO(), userId)
	assert.NoError(t, err)
	assert.Equal(t, sql.NullInt64{Int64: userId, Valid: true}, contributor.UserId)
}

func testGetContributorByEmail(t *testing.T, store *SQLStore, orgId int) {
	email := "user@external.org"
	contributor, err := store.GetContributorByEmail(context.TODO(), email)
	assert.NoError(t, err)
	assert.Equal(t, email, contributor.Email)
}

func testGetContributorByOrcid(t *testing.T, store *SQLStore, orgId int) {
	orcid := "0000-0000-0000-1234"
	contributor, err := store.GetContributorByOrcid(context.TODO(), orcid)
	assert.NoError(t, err)
	assert.Equal(t, sql.NullString{String: orcid, Valid: true}, contributor.Orcid)
}
