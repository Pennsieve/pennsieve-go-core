package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/conflictStrategy"
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
		"Test conflict replace":          testConflictReplace,
		"Test conflict replace no-op":    testConflictReplaceNoConflict,
		"Test PrepareReplace":            testPrepareReplace,
		"Test PrepareReplace empty":      testPrepareReplaceEmpty,
		"Test PrepareReplace validation": testPrepareReplaceValidation,
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
	results, failedPackages, err := store.Queries.addPackageByParent(context.Background(), -1, insertParams, nil)
	assert.Empty(t, failedPackages, "All packages should be inserted correctly.")
	assert.NoError(t, err)
	assert.Len(t, results, 5, "Expect to return 5 packages")

	// Test inserting package with existing Name
	testParams = []test.PackageParams{
		{Name: "package_1.txt", ParentId: -1}}

	insertParams = test.GenerateTestPackages(testParams, 1)
	results, failedPackages, err = store.Queries.addPackageByParent(context.Background(), -1, insertParams, nil)
	assert.NoError(t, err)
	assert.Len(t, results, 0, "Expect to not insert package as there is a conflict.")
	assert.Len(t, failedPackages, 1)
	assert.Equal(t, testParams[0].Name, failedPackages[0].Name)

	// Test inserting package with same name to different dataset
	insertParams = test.GenerateTestPackages(testParams, 2)
	results, failedPackages, err = store.Queries.addPackageByParent(context.Background(), -1, insertParams, nil)
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
	results, failedPackages, err := store.Queries.addPackageByParent(context.Background(), result.Id, insertParams, nil)
	assert.Empty(t, failedPackages, "All packages should be inserted correctly.")
	assert.NoError(t, err)
	assert.Len(t, results, 5, "Expect to return 5 packages")

	// TEST PROVIDED PARENT ID DOES NOT MATCH ALL PARENT IDs
	testParams = []test.PackageParams{
		{Name: "package_6.txt", ParentId: result.Id},
	}
	insertParams = test.GenerateTestPackages(testParams, 1)
	results, failedPackages, err = store.Queries.addPackageByParent(context.Background(), -1, insertParams, nil)
	assert.Error(t, err, "Should return an error when parent_id in call does not match parent_id in params.")
	assert.Nil(t, results)
	assert.Nil(t, failedPackages)

	// TEST MIXED PARENT ID SHOULD FAIL
	testParams = []test.PackageParams{
		{Name: "package_1.txt", ParentId: result.Id},
		{Name: "package_2.txt", ParentId: -1},
	}
	insertParams = test.GenerateTestPackages(testParams, 1)
	results, failedPackages, err = store.Queries.addPackageByParent(context.Background(), result.Id, insertParams, nil)
	assert.Error(t, err, "Should return an error when parent_id in call does not match parent_id in params.")
	assert.Nil(t, results)
	assert.Nil(t, failedPackages)

	// TEST NAMING CONFLICT
	testParams = []test.PackageParams{
		{Name: "package_1.txt", ParentId: result.Id},
	}
	insertParams = test.GenerateTestPackages(testParams, 1)
	results, failedPackages, err = store.Queries.addPackageByParent(context.Background(), result.Id, insertParams, nil)
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
	results, _, err := store.Queries.addPackageByParent(context.Background(), folder2.Id, insertParams, nil)
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
	results, _, err = store.Queries.addPackageByParent(context.Background(), -1, insertParams, nil)
	assert.NoError(t, err)

	ancestorIds, err = store.Queries.GetPackageAncestorIds(context.Background(), results[0].Id)
	assert.NoError(t, err)
	assert.Len(t, ancestorIds, 1)
	assert.Equal(t, ancestorIds[0], results[0].Id, "Expecting first value to be the current package id")

}

// testConflictReplace verifies the Replace strategy soft-deletes the
// predecessor (rename + DELETING), inserts the new package with the
// back-reference populated, and sets replaced_by_package_id on the
// predecessor. Storage counts are left alone — process-jobs-service
// decrements them when it handles the delete job.
func testConflictReplace(t *testing.T, store *SQLStore, orgId int) {
	defer test.Truncate(t, store.db, orgId, "packages")
	defer test.Truncate(t, store.db, orgId, "package_storage")
	defer test.Truncate(t, store.db, orgId, "dataset_storage")

	datasetId := int64(1)

	// Seed: a single package at root.
	original := test.GenerateTestPackages([]test.PackageParams{
		{Name: "report.csv", ParentId: -1},
	}, int(datasetId))
	originalResult, err := store.AddPackagesWithConflict(context.Background(), original, conflictStrategy.KeepBoth)
	assert.NoError(t, err)
	assert.Len(t, originalResult, 1)
	originalId := originalResult[0].Id

	// Seed storage rows so we can verify Replace leaves them alone.
	const predecessorSize = int64(1000)
	assert.NoError(t, store.Queries.IncrementPackageStorage(context.Background(), originalId, predecessorSize))
	assert.NoError(t, store.Queries.IncrementDatasetStorage(context.Background(), datasetId, predecessorSize))

	// Replace: upload a new package with the same name.
	replacement := test.GenerateTestPackages([]test.PackageParams{
		{Name: "report.csv", ParentId: -1},
	}, int(datasetId))
	replacementResult, err := store.AddPackagesWithConflict(context.Background(), replacement, conflictStrategy.Replace)
	assert.NoError(t, err)
	assert.Len(t, replacementResult, 1)

	newPkg := replacementResult[0]
	assert.Equal(t, "report.csv", newPkg.Name, "New package keeps the original requested name")
	assert.NotEqual(t, originalId, newPkg.Id, "Replacement should be a new row")
	assert.True(t, newPkg.ReplacesPackageId.Valid, "replaces_package_id should be set")
	assert.Equal(t, originalId, newPkg.ReplacesPackageId.Int64, "replaces_package_id should point at the predecessor")

	// Predecessor should be state=DELETING, name prefixed, and back-referenced.
	selectStmt := fmt.Sprintf(
		"SELECT name, state, replaced_by_package_id FROM \"%d\".packages WHERE id=$1",
		orgId)
	var predecessorName, predecessorState string
	var predecessorReplacedBy sql.NullInt64
	err = store.db.QueryRow(selectStmt, originalId).Scan(&predecessorName, &predecessorState, &predecessorReplacedBy)
	assert.NoError(t, err)
	assert.Equal(t, packageState.Deleting.String(), predecessorState, "Predecessor should be in DELETING state")
	assert.Contains(t, predecessorName, "__DELETED__", "Predecessor should be renamed with __DELETED__ prefix")
	assert.True(t, predecessorReplacedBy.Valid, "replaced_by_package_id should be set")
	assert.Equal(t, newPkg.Id, predecessorReplacedBy.Int64, "Predecessor's replaced_by_package_id should point at the new row")

	// Storage counts should be untouched — the delete consumer decrements
	// them, not the replace insert.
	predecessorStorage, err := store.Queries.GetPackageStorageById(context.Background(), originalId)
	assert.NoError(t, err)
	assert.Equal(t, predecessorSize, predecessorStorage, "Predecessor package_storage should be unchanged")

	datasetStorage, err := store.Queries.GetDatasetStorageById(context.Background(), datasetId)
	assert.NoError(t, err)
	assert.Equal(t, predecessorSize, datasetStorage, "Dataset storage should be unchanged")
}

// testConflictReplaceNoConflict verifies the Replace strategy inserts
// cleanly (no predecessor rename, no back-reference) when no conflict exists.
func testConflictReplaceNoConflict(t *testing.T, store *SQLStore, orgId int) {
	defer test.Truncate(t, store.db, orgId, "packages")

	fresh := test.GenerateTestPackages([]test.PackageParams{
		{Name: "solo.txt", ParentId: -1},
	}, 1)
	result, err := store.AddPackagesWithConflict(context.Background(), fresh, conflictStrategy.Replace)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "solo.txt", result[0].Name)
	assert.False(t, result[0].ReplacesPackageId.Valid, "No conflict = no replaces_package_id")
	assert.False(t, result[0].ReplacedByPackageId.Valid, "Fresh insert should have null replaced_by_package_id")
}

// testPrepareReplace checks that PrepareReplace renames the old row and
// sets it to DELETING — and nothing else. Storage rows should stay put;
// the delete consumer handles those.
func testPrepareReplace(t *testing.T, store *SQLStore, orgId int) {
	defer test.Truncate(t, store.db, orgId, "packages")
	defer test.Truncate(t, store.db, orgId, "package_storage")
	defer test.Truncate(t, store.db, orgId, "dataset_storage")

	datasetId := int64(1)
	ctx := context.Background()

	// Seed two packages so we can confirm only the target one moves.
	seed := test.GenerateTestPackages([]test.PackageParams{
		{Name: "report.csv", ParentId: -1},
		{Name: "untouched.csv", ParentId: -1},
	}, int(datasetId))
	inserted, err := store.AddPackagesWithConflict(ctx, seed, conflictStrategy.KeepBoth)
	assert.NoError(t, err)
	assert.Len(t, inserted, 2)

	var target, bystander pgdb.Package
	for i := range inserted {
		switch inserted[i].Name {
		case "report.csv":
			target = inserted[i]
		case "untouched.csv":
			bystander = inserted[i]
		}
	}
	originalName := target.Name
	originalNodeId := target.NodeId

	// Add storage rows on the target so we can verify PrepareReplace leaves
	// them alone.
	const targetStorage = int64(1000)
	assert.NoError(t, store.Queries.IncrementPackageStorage(ctx, target.Id, targetStorage))
	assert.NoError(t, store.Queries.IncrementDatasetStorage(ctx, datasetId, targetStorage))

	// Run it.
	err = store.Queries.PrepareReplace(ctx, []*pgdb.Package{&target})
	assert.NoError(t, err)

	// The target row should be renamed and set to DELETING.
	selectStmt := fmt.Sprintf(
		"SELECT name, state FROM \"%d\".packages WHERE id=$1", orgId)
	var gotName, gotState string
	assert.NoError(t, store.db.QueryRow(selectStmt, target.Id).Scan(&gotName, &gotState))
	assert.Equal(t, packageState.Deleting.String(), gotState,
		"state should be DELETING")
	assert.Equal(t, fmt.Sprintf("__DELETED__%s_%s", originalNodeId, originalName), gotName,
		"row should be renamed with __DELETED__ prefix")

	// Storage should not have moved.
	pkgStorage, err := store.Queries.GetPackageStorageById(ctx, target.Id)
	assert.NoError(t, err)
	assert.Equal(t, targetStorage, pkgStorage,
		"package storage should not change")
	dsStorage, err := store.Queries.GetDatasetStorageById(ctx, datasetId)
	assert.NoError(t, err)
	assert.Equal(t, targetStorage, dsStorage,
		"dataset storage should not change")

	// The other package should not move.
	var bystanderName, bystanderState string
	assert.NoError(t, store.db.QueryRow(selectStmt, bystander.Id).Scan(&bystanderName, &bystanderState))
	assert.Equal(t, "untouched.csv", bystanderName)
	assert.NotEqual(t, packageState.Deleting.String(), bystanderState)

	// The name slot should be free: re-inserting the same name in the
	// same parent should succeed and keep the original name.
	reinsert := test.GenerateTestPackages([]test.PackageParams{
		{Name: originalName, ParentId: -1},
	}, int(datasetId))
	reResult, err := store.AddPackagesWithConflict(ctx, reinsert, conflictStrategy.KeepBoth)
	assert.NoError(t, err, "name should be free after PrepareReplace")
	assert.Len(t, reResult, 1)
	assert.Equal(t, originalName, reResult[0].Name,
		"new insert should claim the original name (no auto-rename)")
}

// testPrepareReplaceEmpty: nil or empty input is a no-op, not an error.
func testPrepareReplaceEmpty(t *testing.T, store *SQLStore, _ int) {
	assert.NoError(t, store.Queries.PrepareReplace(context.Background(), nil))
	assert.NoError(t, store.Queries.PrepareReplace(context.Background(), []*pgdb.Package{}))
}

// testPrepareReplaceValidation: bad input should error early rather than
// run a SQL update that matches nothing or writes a malformed name.
func testPrepareReplaceValidation(t *testing.T, store *SQLStore, _ int) {
	ctx := context.Background()
	assert.Error(t, store.Queries.PrepareReplace(ctx, []*pgdb.Package{nil}),
		"nil predecessor entry must error")
	assert.Error(t, store.Queries.PrepareReplace(ctx, []*pgdb.Package{{Id: 0, NodeId: "N:package:x", Name: "x"}}),
		"zero Id must error")
	assert.Error(t, store.Queries.PrepareReplace(ctx, []*pgdb.Package{{Id: 1, NodeId: "", Name: "x"}}),
		"empty NodeId must error")
	// Well-formed input, but the id doesn't exist in the table — the
	// UPDATE matches nothing, so PrepareReplace should fail rather than
	// quietly report success.
	assert.Error(t, store.Queries.PrepareReplace(ctx, []*pgdb.Package{{Id: 999999, NodeId: "N:package:missing", Name: "ghost.csv"}}),
		"unknown id must error")
}
