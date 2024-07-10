package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/role"
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
	addTestDataset(db, "Test Dataset - AddOwnerToDataset")
	addTestDataset(db, "Test Dataset - AddViewerToDataset")
	addTestDataset(db, "Test Dataset - AddEditorToDataset")
	addTestDataset(db, "Test Dataset - AddManagerToDataset")
	defer test.Truncate(t, db, orgId, "datasets")

	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Get Dataset by Name":          testGetDatasetByName,
		"Create Dataset":               testCreateDataset,
		"Add Owner to Dataset":         testAddOwnerToDataset,
		"Add Viewer to Dataset":        testAddViewerToDataset,
		"Add Editor to Dataset":        testAddEditorToDataset,
		"Add Manager to Dataset":       testAddManagerToDataset,
		"Unspecified License is Null":  testUnspecifiedLicenseIsNull,
		"Empty String License Is Null": testEmptyStringLicenseIsNull,
		"Create Dataset Release":       testCreateDatasetRelease,
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
		License:                      "Community Data License Agreement – Sharing",
		Tags:                         nil,
		DataUseAgreement:             defaultDataUseAgreement,
	}
	ds, err := store.CreateDataset(context.TODO(), createDatasetParams)
	assert.NoError(t, err)
	assert.Equal(t, createDatasetParams.Name, ds.Name)
}

func addDatasetUserTest(t *testing.T, store *SQLStore, datasetName string, userId int64, role role.Role, expectedLabel string, expectedPermission int64) {
	ds, err := store.GetDatasetByName(context.TODO(), datasetName)
	assert.NoError(t, err)

	user, err := store.GetUserById(context.TODO(), userId)
	assert.NoError(t, err)

	// add user to the dataset
	dsu1, err := store.AddDatasetUser(context.TODO(), ds, user, role)
	assert.NoError(t, err)
	assert.Equal(t, ds.Id, dsu1.DatasetId)
	assert.Equal(t, user.Id, dsu1.UserId)
	assert.Equal(t, expectedLabel, dsu1.Role)
	assert.Equal(t, expectedPermission, dsu1.PermissionBit)

	// get dataset user
	dsu2, err := store.GetDatasetUser(context.TODO(), ds, user)
	assert.NoError(t, err)
	assert.Equal(t, ds.Id, dsu2.DatasetId)
	assert.Equal(t, user.Id, dsu2.UserId)
	assert.Equal(t, expectedLabel, dsu2.Role)
	assert.Equal(t, expectedPermission, dsu1.PermissionBit)
}

func testAddOwnerToDataset(t *testing.T, store *SQLStore, orgId int) {
	datasetName := "Test Dataset - AddOwnerToDataset"
	userId := int64(1003)
	role := role.Owner
	expectedLabel := "owner"
	expectedPermission := int64(32)

	addDatasetUserTest(t,
		store,
		datasetName,
		userId,
		role,
		expectedLabel,
		expectedPermission)
}

func testAddViewerToDataset(t *testing.T, store *SQLStore, orgId int) {
	datasetName := "Test Dataset - AddViewerToDataset"
	userId := int64(1003)
	role := role.Viewer
	expectedLabel := "viewer"
	expectedPermission := int64(2)

	addDatasetUserTest(t,
		store,
		datasetName,
		userId,
		role,
		expectedLabel,
		expectedPermission)
}

func testAddEditorToDataset(t *testing.T, store *SQLStore, orgId int) {
	datasetName := "Test Dataset - AddEditorToDataset"
	userId := int64(1003)
	role := role.Editor
	expectedLabel := "editor"
	expectedPermission := int64(8)

	addDatasetUserTest(t,
		store,
		datasetName,
		userId,
		role,
		expectedLabel,
		expectedPermission)
}

func testAddManagerToDataset(t *testing.T, store *SQLStore, orgId int) {
	datasetName := "Test Dataset - AddManagerToDataset"
	userId := int64(1003)
	role := role.Manager
	expectedLabel := "manager"
	expectedPermission := int64(16)

	addDatasetUserTest(t,
		store,
		datasetName,
		userId,
		role,
		expectedLabel,
		expectedPermission)
}

func testUnspecifiedLicenseIsNull(t *testing.T, store *SQLStore, orgId int) {
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
		Name:                         "Test Dataset - UnspecifiedLicenseIsNull",
		Description:                  "Test Dataset - UnspecifiedLicenseIsNull",
		Status:                       defaultDatasetStatus,
		AutomaticallyProcessPackages: false,
		Tags:                         nil,
		DataUseAgreement:             defaultDataUseAgreement,
	}
	ds, err := store.CreateDataset(context.TODO(), createDatasetParams)
	assert.NoError(t, err)
	assert.Equal(t, createDatasetParams.Name, ds.Name)
	assert.False(t, ds.License.Valid)
}

func testEmptyStringLicenseIsNull(t *testing.T, store *SQLStore, orgId int) {
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
		Name:                         "Test Dataset - EmptyStringLicenseIsNull",
		Description:                  "Test Dataset - EmptyStringLicenseIsNull",
		Status:                       defaultDatasetStatus,
		AutomaticallyProcessPackages: false,
		License:                      "",
		Tags:                         nil,
		DataUseAgreement:             defaultDataUseAgreement,
	}
	ds, err := store.CreateDataset(context.TODO(), createDatasetParams)
	assert.NoError(t, err)
	assert.Equal(t, createDatasetParams.Name, ds.Name)
	assert.False(t, ds.License.Valid)
}

func testCreateDatasetRelease(t *testing.T, store *SQLStore, orgId int) {
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
		Name:                         "Test Dataset - EmptyStringLicenseIsNull",
		Description:                  "Test Dataset - EmptyStringLicenseIsNull",
		Status:                       defaultDatasetStatus,
		AutomaticallyProcessPackages: false,
		License:                      "",
		Tags:                         nil,
		DataUseAgreement:             defaultDataUseAgreement,
	}
	origin := "GitHub"
	url := "https://github.com/Pennsieve/github-service"
	dsr, err := store.CreateDatasetTypeRelease(context.TODO(), createDatasetParams, origin, url)
	assert.NoError(t, err)
	assert.Equal(t, createDatasetParams.Name, dsr.Dataset.Name)
	assert.Equal(t, origin, dsr.Release.Origin)
	assert.Equal(t, url, dsr.Release.Url)
	assert.Equal(t, dsr.Dataset.Id, dsr.Release.DatasetId)
}
