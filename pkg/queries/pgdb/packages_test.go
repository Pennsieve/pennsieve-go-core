package pgdb

import (
	"context"
	"database/sql"
	"fmt"
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
		"Add package":                       testAddPackage,
		"Test package attributes values":    testPackageAttributeValueAndScan,
		"Test package with duplicate names": testFailDuplicateNames,
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

func testFailDuplicateNames(t *testing.T, store *SQLStore, orgId int) {
	pkgAttr1 := packageInfo.PackageAttribute{
		Key:      "subtype",
		Fixed:    false,
		Value:    "Image",
		Hidden:   false,
		Category: "Pennsieve",
		DataType: "string",
	}
	pkgAttr2 := packageInfo.PackageAttribute{
		Key:      "subtype",
		Fixed:    false,
		Value:    "Image",
		Hidden:   false,
		Category: "Pennsieve",
		DataType: "string",
	}

	records := []pgdb.PackageParams{
		{
			Name:         "folder",
			PackageType:  packageType.Collection,
			PackageState: packageState.Uploaded,
			NodeId:       "N:package:12345678-2222-2222-2222-123456789ABC",
			ParentId:     -1,
			DatasetId:    1,
			OwnerId:      1,
			Size:         123456789,
			ImportId: sql.NullString{
				String: "12345678-2222-2222-2222-123456789ABC",
				Valid:  true,
			},
			Attributes: []packageInfo.PackageAttribute{pkgAttr1, pkgAttr2},
		},
	}
	initialResult, err1 := store.AddPackages(context.Background(), records)
	fmt.Println(initialResult)

	records[0].ImportId = sql.NullString{
		String: "",
		Valid:  false,
	}
	originalNodeID := records[0].NodeId
	records[0].NodeId = "N:package:DUPLICATE-1111-1111-1111-123456789ABC"

	duplicatedResult, err2 := store.AddPackages(context.Background(), records)
	fmt.Println(duplicatedResult)

	assert.Equal(t, err1, nil)
	assert.Equal(t, err2, nil)

	assert.Equal(t, len(initialResult), 1)
	assert.Equal(t, len(duplicatedResult), 1)

	// Check that we get the original node id back
	assert.Equal(t, duplicatedResult[0].NodeId, originalNodeID)
}

// HELPER FUNCTIONS
func truncate(t *testing.T, db *sql.DB, orgID int, table string) {
	query := fmt.Sprintf("TRUNCATE TABLE \"%d\".%s CASCADE", orgID, table)
	_, err := db.Exec(query)
	assert.NoError(t, err)
}
