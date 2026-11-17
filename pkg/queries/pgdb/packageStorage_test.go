package pgdb

import (
	"context"
	"database/sql"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPackageStorage(t *testing.T) {

	orgId := 1
	store := NewSQLStore(testDB[orgId])

	defer func() {
		test.Truncate(t, store.db, orgId, "packages")
		test.Truncate(t, store.db, orgId, "package_storage")
	}()

	// Add Package to test storage on
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
	inserted, err := store.AddPackages(context.Background(), records)
	assert.NoError(t, err)
	if !assert.Len(t, inserted, 1) {
		return
	}
	// Use the id the database assigned: it depends on how many packages earlier
	// tests created, so it cannot be hard-coded.
	packageId := inserted[0].Id

	// Adding 10
	err = store.IncrementPackageStorage(context.Background(), packageId, 10)
	assert.NoError(t, err)

	actualSize, err := store.GetPackageStorageById(context.Background(), packageId)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), actualSize, "Size is expected to be 10")

	// Adding 10
	err = store.IncrementPackageStorage(context.Background(), packageId, 10)
	assert.NoError(t, err)

	actualSize, err = store.GetPackageStorageById(context.Background(), packageId)
	assert.NoError(t, err)
	assert.Equal(t, int64(20), actualSize, "Size is expected to be 20")

	// Removing 10
	err = store.IncrementPackageStorage(context.Background(), packageId, -10)
	assert.NoError(t, err)

	actualSize, err = store.GetPackageStorageById(context.Background(), packageId)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), actualSize, "Size is expected to be 10")

}
