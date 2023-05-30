package pgdb

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/objectType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFiles(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		//"Add file": testAddFile,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := 1
			store := NewSQLStore(testDB[orgId])
			fn(t, store, orgId)
		})
	}
}

func testAddFile(t *testing.T, store *SQLStore, orgId int) {

	defer func() {
		test.Truncate(t, store.db, orgId, "files")
	}()

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
	_, err := store.AddPackages(context.Background(), records)
	assert.NoError(t, err)

	fileUUID := uuid.New()

	files := []pgdb.FileParams{
		{
			PackageId:  1,
			Name:       "file_1.txt",
			FileType:   fileType.MP4,
			S3Bucket:   "storage_bucket",
			S3Key:      "123/123",
			ObjectType: objectType.Source,
			Size:       123,
			CheckSum:   "123213",
			Sha256:     "123213",
			UUID:       fileUUID,
		},
	}

	// Create new file
	results, err := store.AddFiles(context.Background(), files)
	assert.NoError(t, err)
	assert.Equal(t, files[0].Name, results[0].Name)

	// Add file with duplicate UUID --> should return existing file
	files2 := []pgdb.FileParams{
		{
			PackageId:  1,
			Name:       "file_2.txt",
			FileType:   fileType.MP4,
			S3Bucket:   "storage_bucket",
			S3Key:      "123/123",
			ObjectType: objectType.Source,
			Size:       123,
			CheckSum:   "123213",
			Sha256:     "123213",
			UUID:       fileUUID,
		},
	}

	results2, err := store.AddFiles(context.Background(), files2)
	assert.NoError(t, err)
	assert.Equal(t, files[0].Name, results2[0].Name)

}
