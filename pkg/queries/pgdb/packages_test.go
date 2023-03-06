package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestPackageTable is the main Test Suite function for Packages.
func TestPackageTable(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Add package":                    testAddPackage,
		"Test package attributes values": testPackageAttributeValueAndScan,
		"Test duplicate file handling":   checkNameExpansion,
		"Test adding folders":            testAddingFolders,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := 1
			store := NewSQLStore(testDB[orgId])
			fn(t, store, orgId)
		})
	}
}

// TESTS
func testAddPackage(t *testing.T, store *SQLStore, orgId int) {

	defer truncate(t, store.db, orgId, "packages")

	attr := []packageInfo.PackageAttribute{
		{
			Key:      "subtype",
			Fixed:    false,
			Value:    "Image",
			Hidden:   true,
			Category: "Pennsieve",
			DataType: "string",
		}, {
			Key:      "icon",
			Fixed:    false,
			Value:    "Microscope",
			Hidden:   true,
			Category: "Pennsieve",
			DataType: "string",
		},
	}

	records := []pgdb.PackageParams{
		{
			Name:         "TestAddPackage.jpg",
			PackageType:  packageType.Image,
			PackageState: packageState.Ready,
			NodeId:       "N:package:12312314",
			ParentId:     -1,
			DatasetId:    1,
			OwnerId:      1,
			Size:         1000,
			ImportId:     sql.NullString{String: "12323243243245678"},
			Attributes:   attr,
		},
	}

	results, err := store.AddPackages(context.Background(), records)
	assert.NoError(t, err)
	assert.Equal(t, records[0].Name, results[0].Name)

}

func testPackageAttributeValueAndScan(t *testing.T, store *SQLStore, orgId int) {
	emptyAttrs := packageInfo.PackageAttributes{}
	nonEmptyAttrs := packageInfo.PackageAttributes{
		{Key: "subtype",
			Fixed:    false,
			Value:    "Image",
			Hidden:   true,
			Category: "Pennsieve",
			DataType: "string"},
		{Key: "icon",
			Fixed:    false,
			Value:    "Microscope",
			Hidden:   true,
			Category: "Pennsieve",
			DataType: "string"}}
	tests := map[string]struct {
		input    packageInfo.PackageAttributes
		expected packageInfo.PackageAttributes
	}{
		"non-empty": {nonEmptyAttrs, nonEmptyAttrs},
		// If an insert contains a nil PackageAttributes we want to put empty json array in DB
		"nil":   {nil, emptyAttrs},
		"empty": {emptyAttrs, emptyAttrs},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			p := pgdb.Package{
				Name:         "image.jpg",
				PackageType:  packageType.Image,
				PackageState: packageState.Ready,
				NodeId:       "N:package:1234",
				DatasetId:    1,
				OwnerId:      1,
				Attributes:   data.input}
			insert := fmt.Sprintf(
				"INSERT INTO \"%d\".packages (name, type, state, node_id, dataset_id, owner_id, attributes) VALUES ($1, $2, $3, $4, $5, $6, $7)",
				orgId)
			_, err := store.db.Exec(insert, p.Name, p.PackageType, p.PackageState, p.NodeId, p.DatasetId, p.OwnerId, p.Attributes)
			assert.NoError(t, err)
			defer truncate(t, store.db, orgId, "packages")

			countStmt := fmt.Sprintf("SELECT COUNT(*) FROM \"%d\".packages", orgId)
			var count int
			assert.NoError(t, store.db.QueryRow(countStmt).Scan(&count))
			assert.Equal(t, 1, count)

			selectStmt := fmt.Sprintf(
				"SELECT name, type, state, node_id, dataset_id, owner_id, attributes FROM \"%d\".packages",
				orgId)

			var actual pgdb.Package
			assert.NoError(t, store.db.QueryRow(selectStmt).Scan(
				&actual.Name,
				&actual.PackageType,
				&actual.PackageState,
				&actual.NodeId,
				&actual.DatasetId,
				&actual.OwnerId,
				&actual.Attributes))

			assert.Equal(t, p.Name, actual.Name)
			assert.Equal(t, p.PackageType, actual.PackageType)
			assert.Equal(t, p.PackageState, actual.PackageState)
			assert.Equal(t, p.NodeId, actual.NodeId)
			assert.Equal(t, p.DatasetId, actual.DatasetId)
			assert.Equal(t, p.OwnerId, actual.OwnerId)
			assert.Equal(t, data.expected, actual.Attributes)
		})
	}
}

func checkNameExpansion(t *testing.T, _ *SQLStore, _ int) {
	existingNames := []string{
		"file1.doc",
		"file2.doc",
		"file3.doc",
		"file2 (2).doc",
	}

	// Check new file
	currentPackage := pgdb.PackageParams{
		Name: "new_file.doc",
	}
	checkUpdateName(&currentPackage, 1, "", existingNames)
	assert.Equal(t, "new_file.doc", currentPackage.Name, "Non existing file should remain unchanged")

	currentPackage = pgdb.PackageParams{
		Name: "file1.doc",
	}
	checkUpdateName(&currentPackage, 1, "", existingNames)
	assert.Equal(t, "file1 (2).doc", currentPackage.Name, "File with existing name should be appended with (2)")

	currentPackage = pgdb.PackageParams{
		Name: "file2.doc",
	}
	checkUpdateName(&currentPackage, 1, "", existingNames)
	assert.Equal(t, "file2 (3).doc", currentPackage.Name, "File with existing appended name should be have index increased (3)")

}

func testAddingFolders(t *testing.T, store *SQLStore, orgId int) {
	defer truncate(t, store.db, orgId, "packages")

	// TEST ADDING FOLDERS TO ROOT
	uploadId, _ := uuid.NewUUID()
	folder := pgdb.PackageParams{
		Name:         "Folder1",
		PackageType:  packageType.Collection,
		PackageState: packageState.Ready,
		NodeId:       fmt.Sprintf("N:Package:%s", uploadId.String()),
		ParentId:     -1,
		DatasetId:    1,
		OwnerId:      1,
		Size:         1000, // should be ignored
		ImportId:     sql.NullString{String: uploadId.String(), Valid: true},
		Attributes:   []packageInfo.PackageAttribute{},
	}

	result, err := store.Queries.AddFolder(context.Background(), folder)
	assert.NoError(t, err)
	assert.Equal(t, folder.Name, result.Name, "name of resulting folder should be correct.")
	assert.False(t, result.ParentId.Valid, "should not have a parent id.")
	assert.False(t, result.Size.Valid, "folder size should be nil.")

	uploadId, _ = uuid.NewUUID()
	folder2 := pgdb.PackageParams{
		Name:         "Folder2",
		PackageType:  packageType.Collection,
		PackageState: packageState.Ready,
		NodeId:       fmt.Sprintf("N:Package:%s", uploadId.String()),
		ParentId:     -1,
		DatasetId:    1,
		OwnerId:      1,
		Size:         1000,
		ImportId:     sql.NullString{String: uploadId.String(), Valid: true},
		Attributes:   []packageInfo.PackageAttribute{},
	}
	result2, err := store.Queries.AddFolder(context.Background(), folder2)
	assert.NoError(t, err)
	assert.Equal(t, folder2.Name, result2.Name)
	assert.NotEqualf(t, result.Id, result2.Id, "Adding two folders should return object with different IDs")

	result3, err := store.Queries.AddFolder(context.Background(), folder)
	assert.Equal(t, folder.Name, result.Name)
	assert.Equal(t, result.Id, result3.Id, "conflict should return the existing folder")

	uploadId, _ = uuid.NewUUID()
	badFolder := pgdb.PackageParams{
		Name:         "Image",
		PackageType:  packageType.Image,
		PackageState: packageState.Ready,
		NodeId:       fmt.Sprintf("N:Package:%s", uploadId.String()),
		ParentId:     -1,
		DatasetId:    1,
		OwnerId:      1,
		Size:         1000,
		ImportId:     sql.NullString{String: uploadId.String(), Valid: true},
		Attributes:   []packageInfo.PackageAttribute{},
	}
	result4, err := store.Queries.AddFolder(context.Background(), badFolder)
	assert.Error(t, err, "Adding folder while specifying non-collection package should error")
	assert.Nil(t, result4, "Adding non-folder using addfolder method should return nil")

	// TEST ADDING FOLDERS TO EXISTING FOLDER
	uploadId, _ = uuid.NewUUID()
	nestedFolder1 := pgdb.PackageParams{
		Name:         "NestedFolder1",
		PackageType:  packageType.Collection,
		PackageState: packageState.Ready,
		NodeId:       fmt.Sprintf("N:Package:%s", uploadId.String()),
		ParentId:     result.Id,
		DatasetId:    1,
		OwnerId:      1,
		Size:         1000,
		ImportId:     sql.NullString{String: uploadId.String(), Valid: true},
		Attributes:   []packageInfo.PackageAttribute{},
	}
	result5, err := store.Queries.AddFolder(context.Background(), nestedFolder1)
	assert.NoError(t, err)
	assert.Equal(t, nestedFolder1.Name, result5.Name)
	assert.True(t, result5.ParentId.Valid, "Package should hava a parent id")
	resultParentId, _ := result5.ParentId.Value()
	assert.Equal(t, result.Id, resultParentId, "Parent ID should be ID of parent package")

	uploadId, _ = uuid.NewUUID()
	nestedFolder2 := pgdb.PackageParams{
		Name:         "NestedFolder1",
		PackageType:  packageType.Collection,
		PackageState: packageState.Ready,
		NodeId:       fmt.Sprintf("N:Package:%s", uploadId.String()),
		ParentId:     result.Id,
		DatasetId:    1,
		OwnerId:      1,
		Size:         1000,
		ImportId:     sql.NullString{String: uploadId.String(), Valid: true},
		Attributes:   []packageInfo.PackageAttribute{},
	}

	// TEST ADDING NESTED FOLDER WITH SAME NAME
	result6, err := store.Queries.AddFolder(context.Background(), nestedFolder2)
	assert.Equal(t, result5.Id, result6.Id, "conflict should return the existing folder")

}

// HELPER FUNCTIONS
func truncate(t *testing.T, db *sql.DB, orgID int, table string) {
	query := fmt.Sprintf("TRUNCATE TABLE \"%d\".%s CASCADE", orgID, table)
	_, err := db.Exec(query)
	assert.NoError(t, err)
}

//
//func testNestedPackages(t *testing.T, db *sql.DB, orgID int, table string) {
//
//	ctx := context.Background()
//
//	records := []pgdb.PackageParams{
//		{
//			Name:         "",
//			PackageType:  0,
//			PackageState: 0,
//			NodeId:       "",
//			ParentId:     0,
//			DatasetId:    0,
//			OwnerId:      0,
//			Size:         0,
//			ImportId:     sql.NullString{},
//			Attributes:   nil,
//		},
//
//	}
//	AddPackages(ctx context.Context, records []pgdb.PackageParams)
//
//
//}
