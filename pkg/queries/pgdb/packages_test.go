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
	"github.com/pennsieve/pennsieve-go-core/test"
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
		"Test adding folders":            testAddingFolders,
		"Test adding packages to root":   testAddingPackagesToRoot,
		"Test adding nested packages":    testAddingNestedPackages,
		"Test name expansion":            testNameExpansion,
		"Test getting ancestor Ids":      testGettingAncestors,
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

	defer test.Truncate(t, store.db, orgId, "packages")

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
			defer test.Truncate(t, store.db, orgId, "packages")

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

//testAddingFolders tests adding folders to datasets
func testAddingFolders(t *testing.T, store *SQLStore, orgId int) {
	defer test.Truncate(t, store.db, orgId, "packages")

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
	assert.Equal(t, folder2.NodeId, result2.NodeId, "Node Id should match added package node id")

	result3, err := store.Queries.AddFolder(context.Background(), folder)
	assert.Equal(t, folder.Name, result.Name)
	assert.Equal(t, result.Id, result3.Id, "conflict should return the existing folder")
	assert.Equal(t, result.NodeId, result3.NodeId, "conflict should return the existing folder")

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

func testAddingPackagesToRoot(t *testing.T, store *SQLStore, orgId int) {
	defer test.Truncate(t, store.db, orgId, "packages")

	// Test adding packages to root
	testParams := []test.PackageParams{
		{Name: "package_1.txt", ParentId: -1},
		{Name: "package_2.txt", ParentId: -1},
		{Name: "package_3.txt", ParentId: -1},
		{Name: "package_4.txt", ParentId: -1},
		{Name: "package_5.txt", ParentId: -1},
	}

	insertParams := test.GenerateTestPackages(testParams, 1)
	results, failedPackages, err := store.Queries.addPackageByParent(context.Background(), -1, insertParams)
	assert.Empty(t, failedPackages, "All packages should be inserted correctly.")
	assert.NoError(t, err)
	assert.Len(t, results, 5, "Expect to return 5 packages")

	// Test inserting package with existing Name
	testParams = []test.PackageParams{
		{Name: "package_1.txt", ParentId: -1}}

	insertParams = test.GenerateTestPackages(testParams, 1)
	results, failedPackages, err = store.Queries.addPackageByParent(context.Background(), -1, insertParams)
	assert.NoError(t, err)
	assert.Len(t, results, 0, "Expect to not insert package as there is a conflict.")
	assert.Len(t, failedPackages, 1)
	assert.Equal(t, testParams[0].Name, failedPackages[0].Name)

	// Test inserting package with same name to different dataset
	insertParams = test.GenerateTestPackages(testParams, 2)
	results, failedPackages, err = store.Queries.addPackageByParent(context.Background(), -1, insertParams)
	assert.NoError(t, err)
	assert.Len(t, results, 1, "Expect to insert package in dataset 2.")
	assert.Len(t, failedPackages, 0)
	assert.Equal(t, testParams[0].Name, results[0].Name)
}

func testAddingNestedPackages(t *testing.T, store *SQLStore, orgId int) {
	defer test.Truncate(t, store.db, orgId, "packages")

	// ADD FOLDER TO ROOT
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

	// ADD NESTED FOLDER
	uploadId, _ = uuid.NewUUID()
	folder = pgdb.PackageParams{
		Name:         "Folder1",
		PackageType:  packageType.Collection,
		PackageState: packageState.Ready,
		NodeId:       fmt.Sprintf("N:Package:%s", uploadId.String()),
		ParentId:     result.Id,
		DatasetId:    1,
		OwnerId:      1,
		Size:         1000, // should be ignored
		ImportId:     sql.NullString{String: uploadId.String(), Valid: true},
		Attributes:   []packageInfo.PackageAttribute{},
	}

	result, err = store.Queries.AddFolder(context.Background(), folder)
	assert.NoError(t, err)

	// Test adding packages to root
	testParams := []test.PackageParams{
		{Name: "package_1.txt", ParentId: result.Id},
		{Name: "package_2.txt", ParentId: result.Id},
		{Name: "package_3.txt", ParentId: result.Id},
		{Name: "package_4.txt", ParentId: result.Id},
		{Name: "package_5.txt", ParentId: result.Id},
	}

	insertParams := test.GenerateTestPackages(testParams, 1)
	results, failedPackages, err := store.Queries.addPackageByParent(context.Background(), result.Id, insertParams)
	assert.Empty(t, failedPackages, "All packages should be inserted correctly.")
	assert.NoError(t, err)
	assert.Len(t, results, 5, "Expect to return 5 packages")

	// TEST PROVIDED PARENT ID DOES NOT MATCH ALL PARENT IDs
	testParams = []test.PackageParams{
		{Name: "package_6.txt", ParentId: result.Id},
	}
	insertParams = test.GenerateTestPackages(testParams, 1)
	results, failedPackages, err = store.Queries.addPackageByParent(context.Background(), -1, insertParams)
	assert.Error(t, err, "Should return an error when parent_id in call does not match parent_id in params.")
	assert.Nil(t, results)
	assert.Nil(t, failedPackages)

	// TEST MIXED PARENT ID SHOULD FAIL
	testParams = []test.PackageParams{
		{Name: "package_1.txt", ParentId: result.Id},
		{Name: "package_2.txt", ParentId: -1},
	}
	insertParams = test.GenerateTestPackages(testParams, 1)
	results, failedPackages, err = store.Queries.addPackageByParent(context.Background(), result.Id, insertParams)
	assert.Error(t, err, "Should return an error when parent_id in call does not match parent_id in params.")
	assert.Nil(t, results)
	assert.Nil(t, failedPackages)

	// TEST NAMING CONFLICT
	testParams = []test.PackageParams{
		{Name: "package_1.txt", ParentId: result.Id},
	}
	insertParams = test.GenerateTestPackages(testParams, 1)
	results, failedPackages, err = store.Queries.addPackageByParent(context.Background(), result.Id, insertParams)
	assert.NoError(t, err)
	assert.Len(t, results, 0, "Expect to not insert package as there is a naming conflict.")
	assert.Len(t, failedPackages, 1, "Expect package to fail.")

}

func testAddingMixedParentPackages(t *testing.T, store *SQLStore, orgId int) {
	defer test.Truncate(t, store.db, orgId, "packages")

	// ADD FOLDER TO ROOT
	uploadId, _ := uuid.NewUUID()
	folderParams := pgdb.PackageParams{
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

	folder1, err := store.Queries.AddFolder(context.Background(), folderParams)
	assert.NoError(t, err)

	// ADD NESTED FOLDER
	uploadId, _ = uuid.NewUUID()
	folderParams = pgdb.PackageParams{
		Name:         "Folder2",
		PackageType:  packageType.Collection,
		PackageState: packageState.Ready,
		NodeId:       fmt.Sprintf("N:Package:%s", uploadId.String()),
		ParentId:     folder1.Id,
		DatasetId:    1,
		OwnerId:      1,
		Size:         1000, // should be ignored
		ImportId:     sql.NullString{String: uploadId.String(), Valid: true},
		Attributes:   []packageInfo.PackageAttribute{},
	}

	folder2, err := store.Queries.AddFolder(context.Background(), folderParams)
	assert.NoError(t, err)

	// Test adding packages to root
	testParams := []test.PackageParams{
		{Name: "package_1.txt", ParentId: -1},
		{Name: "package_2.txt", ParentId: -1},
		{Name: "package_3.txt", ParentId: folder1.Id},
		{Name: "package_4.txt", ParentId: folder2.Id},
		{Name: "package_5.txt", ParentId: folder2.Id},
		{Name: "package_5.txt", ParentId: folder2.Id},
		{Name: "package_5.txt", ParentId: folder2.Id},
	}

	insertParams := test.GenerateTestPackages(testParams, 1)
	results, err := store.AddPackages(context.Background(), insertParams)
	assert.NoError(t, err)
	assert.Len(t, results, 7, "Expect to return 5 packages")
	for _, p := range results {
		switch p.NodeId {
		case insertParams[0].NodeId:
			assert.False(t, p.ParentId.Valid)
		case insertParams[1].NodeId:
			assert.False(t, p.ParentId.Valid)
		case insertParams[2].NodeId:
			actualParentId, _ := p.ParentId.Value()
			assert.Equal(t, folder1.Id, actualParentId)
		case insertParams[3].NodeId:
			actualParentId, _ := p.ParentId.Value()
			assert.Equal(t, folder2.Id, actualParentId)
		case insertParams[4].NodeId:
			actualParentId, _ := p.ParentId.Value()
			assert.Equal(t, folder2.Id, actualParentId)
		case insertParams[5].NodeId:
			actualParentId, _ := p.ParentId.Value()
			assert.Equal(t, folder2.Id, actualParentId)
			assert.Equal(t, "package_5 (1).txt", p.Name, "First duplicate Name in same addFiles request should be appended by (1).")
		case insertParams[6].NodeId:
			actualParentId, _ := p.ParentId.Value()
			assert.Equal(t, folder2.Id, actualParentId)
			assert.Equal(t, "package_5 (2).txt", p.Name, "Second duplicate Name in same addFiles request should be appended by (2).")
		default:
			assert.Fail(t, "Returned unexpected package")
		}
	}

	// TEST ADDING DOUBLE DUPLICATE
	testParams = []test.PackageParams{
		{Name: "package_5.txt", ParentId: folder2.Id},
	}
	insertParams = test.GenerateTestPackages(testParams, 1)
	results, err = store.AddPackages(context.Background(), insertParams)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	for _, p := range results {
		switch p.NodeId {
		case insertParams[0].NodeId:
			actualParentId, _ := p.ParentId.Value()
			assert.Equal(t, folder2.Id, actualParentId)
			assert.Equal(t, "package_5 (3).txt", p.Name, "Third duplicate Name in separate addFiles request should be appended by (3).")
		default:
			assert.Fail(t, "Did not find expected package")
		}
	}
}

func testNameExpansion(t *testing.T, _ *SQLStore, _ int) {

	originalName := "file.doc"

	// Check new file
	currentPackage := pgdb.PackageParams{
		Name: "file.doc",
	}
	expandName(&currentPackage, originalName, 1)
	assert.Equal(t, "file (1).doc", currentPackage.Name, "File with existing name should be appended with (1)")

	currentPackage = pgdb.PackageParams{
		Name: "file (2).doc",
	}
	expandName(&currentPackage, originalName, 3)
	assert.Equal(t, "file (3).doc", currentPackage.Name, "File with existing appended name should be have index increased (3)")

	// Test name with multiple periods
	originalName = "file.gz.tar"
	currentPackage = pgdb.PackageParams{
		Name: "file.gz.tar",
	}
	expandName(&currentPackage, originalName, 1)
	assert.Equal(t, "file (1).gz.tar", currentPackage.Name, "File with existing name should be appended with (1)")

	// File without extension
	originalName = "file"
	currentPackage = pgdb.PackageParams{
		Name: "file",
	}
	expandName(&currentPackage, originalName, 1)
	assert.Equal(t, "file (1)", currentPackage.Name, "File with existing name should be appended with (1)")

	// File without extension
	originalName = "file"
	currentPackage = pgdb.PackageParams{
		Name: "file (1)",
	}
	expandName(&currentPackage, originalName, 2)
	assert.Equal(t, "file (2)", currentPackage.Name, "File with existing name should be appended with (2)")

	// File with spaces
	originalName = "file one.txt"
	currentPackage = pgdb.PackageParams{
		Name: "file one.txt",
	}
	expandName(&currentPackage, originalName, 1)
	assert.Equal(t, "file one (1).txt", currentPackage.Name, "File with existing name should be appended with (2)")

	// File with spaces
	originalName = "file one.txt"
	currentPackage = pgdb.PackageParams{
		Name: "file one (1).txt",
	}
	expandName(&currentPackage, originalName, 2)
	assert.Equal(t, "file one (2).txt", currentPackage.Name, "File with existing name should be appended with (2)")
}

// testGettingAncestors tests that getAncestors method returns list of parent folders.
func testGettingAncestors(t *testing.T, store *SQLStore, orgId int) {
	defer test.Truncate(t, store.db, orgId, "packages")

	// ADD FOLDER TO ROOT
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

	folder1, err := store.Queries.AddFolder(context.Background(), folder)
	assert.NoError(t, err)

	// ADD NESTED FOLDER
	uploadId, _ = uuid.NewUUID()
	folder = pgdb.PackageParams{
		Name:         "Folder1",
		PackageType:  packageType.Collection,
		PackageState: packageState.Ready,
		NodeId:       fmt.Sprintf("N:Package:%s", uploadId.String()),
		ParentId:     folder1.Id,
		DatasetId:    1,
		OwnerId:      1,
		Size:         1000, // should be ignored
		ImportId:     sql.NullString{String: uploadId.String(), Valid: true},
		Attributes:   []packageInfo.PackageAttribute{},
	}

	folder2, err := store.Queries.AddFolder(context.Background(), folder)
	assert.NoError(t, err)

	// Test adding packages to root
	testParams := []test.PackageParams{
		{Name: "package_1.txt", ParentId: folder2.Id},
	}
	insertParams := test.GenerateTestPackages(testParams, 1)
	results, _, err := store.Queries.addPackageByParent(context.Background(), folder2.Id, insertParams)
	assert.NoError(t, err)

	ancestorIds, err := store.Queries.GetPackageAncestorIds(context.Background(), results[0].Id)
	assert.NoError(t, err)
	assert.Len(t, ancestorIds, 3)
	assert.Equal(t, ancestorIds[0], results[0].Id, "Expecting first value to be the current package id")
	assert.Equal(t, ancestorIds[1], folder2.Id, "Expecting first value to be folder for package")
	assert.Equal(t, ancestorIds[2], folder1.Id, "Expecting second value to be the first folder in the root")

	// ---- Test Add package to Root ----
	testParams = []test.PackageParams{
		{Name: "package_root.txt", ParentId: -1},
	}
	insertParams = test.GenerateTestPackages(testParams, 1)
	results, _, err = store.Queries.addPackageByParent(context.Background(), -1, insertParams)
	assert.NoError(t, err)

	ancestorIds, err = store.Queries.GetPackageAncestorIds(context.Background(), results[0].Id)
	assert.NoError(t, err)
	assert.Len(t, ancestorIds, 1)
	assert.Equal(t, ancestorIds[0], results[0].Id, "Expecting first value to be the current package id")

}
