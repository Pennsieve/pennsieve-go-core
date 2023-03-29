package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

//goland:noinspection SqlResolve
func TestDatasetsInsertSelect(t *testing.T) {
	orgId := 3
	db := testDB[orgId]
	defer test.Truncate(t, testDB[orgId], orgId, "datasets")

	input := pgdb.Dataset{
		Id:           1000,
		Name:         "Test Dataset",
		State:        "READY",
		Description:  sql.NullString{},
		NodeId:       sql.NullString{String: "N:dataset:1234", Valid: true},
		Role:         sql.NullString{String: "editor", Valid: true},
		Tags:         pgdb.Tags{"test", "sql"},
		Contributors: pgdb.Contributors{},
		StatusId:     int32(1),
	}
	_, err := db.Exec("INSERT INTO datasets (id, name, state, description, node_id, role, tags, contributors, status_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", input.Id, input.Name, input.State, input.Description, input.NodeId, input.Role, input.Tags, input.Contributors, input.StatusId)

	if assert.NoError(t, err) {

		countStmt := fmt.Sprintf("SELECT COUNT(*) FROM datasets")
		var count int
		assert.NoError(t, db.QueryRow(countStmt).Scan(&count))
		assert.Equal(t, 1, count)

		var actual pgdb.Dataset
		err = db.QueryRow("SELECT id, name, state, description, node_id, role, tags, contributors, status_id FROM datasets").Scan(
			&actual.Id,
			&actual.Name,
			&actual.State,
			&actual.Description,
			&actual.NodeId,
			&actual.Role,
			&actual.Tags,
			&actual.Contributors,
			&actual.StatusId)
		if assert.NoError(t, err) {
			assert.Equal(t, input.Name, actual.Name)
			assert.Equal(t, input.State, actual.State)
			assert.Equal(t, input.NodeId, actual.NodeId)
			assert.Equal(t, input.Role, actual.Role)
			assert.Equal(t, input.StatusId, actual.StatusId)

			assert.Equal(t, input.Tags, actual.Tags)
			assert.Equal(t, input.Contributors, actual.Contributors)
			assert.False(t, actual.Description.Valid)
		}
	}

}

func TestDatasets(t *testing.T) {
	orgId := 3
	db := testDB[orgId]
	store := NewSQLStore(db)

	addTestDataset(db, "Test Dataset - GetDatasetByName")
	addTestDataset(db, "Test Dataset - AddUserToDataset")
	defer test.Truncate(t, db, orgId, "datasets")

	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Get Dataset by Name": testGetDatasetByName,
		"Create Dataset":      testCreateDataset,
		"Add User to Dataset": testAddUserToDataset,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := orgId
			store := store
			fn(t, store, orgId)
		})
	}
}

func testGetDatasetByName(t *testing.T, store *SQLStore, orgId int) {
	name := "Test Dataset - GetDatasetByName"
	ds, err := store.GetDatasetByName(context.TODO(), name)
	assert.NoError(t, err)
	assert.Equal(t, name, ds.Name)
}

func testCreateDataset(t *testing.T, store *SQLStore, orgId int) {
	var err error
	defaultDatasetStatus, err := store.GetDefaultDatasetStatus(context.TODO(), orgId)
	if err != nil {
		fmt.Errorf("testCreateDataset(): failed to get default dataset status")
	}
	defaultDataUseAgreement, err := store.GetDefaultDataUseAgreement(context.TODO(), orgId)
	if err != nil {
		fmt.Errorf("testCreateDataset(): failed to get default data use agreement")
	}
	createDatasetParams := CreateDatasetParams{
		Name:                         "Test Dataset - CreateDataset",
		Description:                  "Test Dataset - CreateDataset",
		Status:                       defaultDatasetStatus,
		AutomaticallyProcessPackages: false,
		License:                      "",
		Tags:                         nil,
		DataUseAgreement:             defaultDataUseAgreement,
	}
	ds, err := store.CreateDataset(context.TODO(), createDatasetParams)
	assert.NoError(t, err)
	assert.Equal(t, createDatasetParams.Name, ds.Name)
}

func testAddUserToDataset(t *testing.T, store *SQLStore, orgId int) {
	ds, err := store.GetDatasetByName(context.TODO(), "Test Dataset - AddUserToDataset")
	assert.NoError(t, err)

	user, err := store.GetUserById(context.TODO(), 1003)
	assert.NoError(t, err)

	// add user to the dataset
	dsu1, err := store.AddDatasetUser(context.TODO(), ds, user, dataset.Owner)
	assert.NoError(t, err)
	assert.Equal(t, ds.Id, dsu1.DatasetId)
	assert.Equal(t, user.Id, dsu1.UserId)
	assert.Equal(t, dataset.Owner, dsu1.Role)

	// get dataset user
	dsu2, err := store.GetDatasetUser(context.TODO(), ds, user)
	assert.NoError(t, err)
	assert.Equal(t, ds.Id, dsu2.DatasetId)
	assert.Equal(t, user.Id, dsu2.UserId)
	assert.Equal(t, dataset.Owner, dsu2.Role)
}
