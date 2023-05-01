package pgdb

import (
	"context"
	"database/sql"
	"errors"
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
		"Add Contributor Existing User":     testAddContributorExistingUser,
		"Add Contributor External":          testAddContributorExternal,
		"Add Contributor With Blank Degree": testAddContributorWithBlankDegree,
		"Add Contributor Without a Degree":  testAddContributorWithoutDegree,
		"Get Contributor By Id":             testGetContributorById,
		"Get Contributor By User Id":        testGetContributorByUserId,
		"Get Contributor By Email":          testGetContributorByEmail,
		"Get Contributor By Orcid":          testGetContributorByOrcid,
		"Find Contributor by User Id":       testFindContributorByUserId,
		"Find Contributor by Email":         testFindContributorByEmail,
		"Find Contributor by Orcid":         testFindContributorByOrcid,
		"Find Non-existent Contributor":     testFindNonexistentContributor,
		"Add Existing Contributor":          testAddExistingContributor,
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

func testAddContributorWithBlankDegree(t *testing.T, store *SQLStore, orgId int) {
	newContributor := NewContributor{
		FirstName:    "Blank",
		LastName:     "Degree",
		Degree:       "",
		EmailAddress: "blank.degree@not-pennsieve.org",
		Orcid:        "0000-0000-0000-5555",
	}

	contributor, err := store.AddContributor(context.TODO(), newContributor)
	assert.NoError(t, err)
	assert.Equal(t, newContributor.EmailAddress, contributor.Email)
	assert.False(t, contributor.Degree.Valid)
}

func testAddContributorWithoutDegree(t *testing.T, store *SQLStore, orgId int) {
	newContributor := NewContributor{
		FirstName:    "Without",
		LastName:     "Degree",
		Degree:       "",
		EmailAddress: "without.degree@not-pennsieve.org",
		Orcid:        "0000-0000-0000-6666",
	}

	contributor, err := store.AddContributor(context.TODO(), newContributor)
	assert.NoError(t, err)
	assert.Equal(t, newContributor.EmailAddress, contributor.Email)
	assert.False(t, contributor.Degree.Valid)
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

func testFindContributorByUserId(t *testing.T, store *SQLStore, orgId int) {
	newContributor := NewContributor{
		UserId: 1004,
	}
	contributor, err := store.FindContributor(context.TODO(), newContributor)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, contributor)
	assert.Equal(t, newContributor.UserId, contributor.UserId.Int64)
}

func testFindContributorByEmail(t *testing.T, store *SQLStore, orgId int) {
	newContributor := NewContributor{
		EmailAddress: "user4@pennsieve.org",
	}
	contributor, err := store.FindContributor(context.TODO(), newContributor)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, contributor)
	assert.Equal(t, newContributor.EmailAddress, contributor.Email)
}

func testFindContributorByOrcid(t *testing.T, store *SQLStore, orgId int) {
	newContributor := NewContributor{
		Orcid: "0000-0000-0000-4444",
	}
	contributor, err := store.FindContributor(context.TODO(), newContributor)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, contributor)
	assert.Equal(t, newContributor.Orcid, contributor.Orcid.String)
}

func testFindNonexistentContributor(t *testing.T, store *SQLStore, orgId int) {
	contributorNotFoundError := &ContributorNotFoundError{}
	newContributor := NewContributor{
		FirstName:    "None",
		LastName:     "None",
		EmailAddress: "none@none.org",
		Orcid:        "9999-9999-9999-9999",
		UserId:       123456,
	}
	contributor, err := store.FindContributor(context.TODO(), newContributor)
	assert.True(t, errors.As(err, contributorNotFoundError))
	assert.Nil(t, contributor)
}

func testAddExistingContributor(t *testing.T, store *SQLStore, orgId int) {
	// get existing contributor by User Id
	userId := int64(1004)
	existing, err := store.GetContributorByUserId(context.TODO(), userId)
	assert.NoError(t, err)
	assert.NotNil(t, existing)

	// create a NewContributor based on the found contributor
	newContributor := NewContributor{
		FirstName:    existing.FirstName,
		LastName:     existing.LastName,
		EmailAddress: existing.Email,
		Orcid:        existing.Orcid.String,
		UserId:       existing.UserId.Int64,
	}

	// try to add the NewContributor
	added, err := store.AddContributor(context.TODO(), newContributor)
	assert.NoError(t, err)
	assert.NotNil(t, added)

	// verify that the existing and added contributors are the same
	assert.Equal(t, existing.Id, added.Id)
}
