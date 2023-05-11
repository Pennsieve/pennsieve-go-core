package pgdb

import (
	"context"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/objectType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFiles(t *testing.T) {
	orgId := 3
	db := testDB[orgId]
	store := NewSQLStore(db)

	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int, packageId int,
	){
		"AddFiles duplicate uuid":                    testAddFilesDuplicateUUID,
		"AddFiles duplicate uuid, differing S3 keys": testAddFilesDuplicateUUIDDifferentS3Key,
	} {

		t.Run(scenario, func(t *testing.T) {
			defer test.Truncate(t, db, orgId, "packages")
			defer test.Truncate(t, db, orgId, "datasets")

			datasetId := addTestDataset(db, "TestFiles Dataset")
			packages, err := store.AddPackages(context.Background(),
				test.GenerateTestPackages([]test.PackageParams{{Name: "test-package", ParentId: -1}}, int(datasetId)))
			if err != nil {
				assert.FailNow(t, "unable to set up test; error inserting package", err)
			}
			if len(packages) != 1 {
				assert.FailNow(t, "unable to set up test; unexpected number of packages", packages)
			}
			packageId := int(packages[0].Id)

			orgId := orgId
			store := store
			fn(t, store, orgId, packageId)
		})
	}
}

func testAddFilesDuplicateUUID(t *testing.T, store *SQLStore, orgId int, packageId int) {
	defer test.Truncate(t, store.db, orgId, "files")

	s3Bucket := "test-bucket"
	s3Key := "test/s3/key/a1b2.edf"
	uuid := uuid.Must(uuid.NewUUID())
	files := []pgdb.FileParams{{
		PackageId:  packageId,
		Name:       "test-file",
		FileType:   fileType.EDF,
		S3Bucket:   s3Bucket,
		S3Key:      s3Key,
		ObjectType: objectType.Source,
		Size:       1024,
		CheckSum:   "",
		Sha256:     "",
		UUID:       uuid,
	}}
	var actualFileId string
	var actualFileUpdatedAt time.Time
	actualFiles, err := store.AddFiles(context.Background(), files)
	if assert.NoError(t, err) {
		assert.Len(t, actualFiles, 1)
		assert.Equal(t, s3Bucket, actualFiles[0].S3Bucket)
		assert.Equal(t, s3Key, actualFiles[0].S3Key)
		assert.Equal(t, uuid, actualFiles[0].UUID)
		actualFileId = actualFiles[0].Id
		assert.NotEmpty(t, actualFileId)
		actualFileUpdatedAt = actualFiles[0].UpdatedAt
		assert.NotEmpty(t, actualFileUpdatedAt)
	}
	duplicateFiles, err := store.AddFiles(context.Background(), files)
	if assert.NoError(t, err) {
		assert.Len(t, duplicateFiles, 1)
		duplicateFile := duplicateFiles[0]
		assert.Equal(t, s3Bucket, duplicateFile.S3Bucket)
		assert.Equal(t, s3Key, duplicateFile.S3Key)
		assert.Equal(t, uuid, duplicateFile.UUID)
		assert.Equal(t, actualFileId, duplicateFile.Id)
		assert.True(t, actualFileUpdatedAt.Before(duplicateFile.UpdatedAt))
	}
}

func testAddFilesDuplicateUUIDDifferentS3Key(t *testing.T, store *SQLStore, orgId int, packageId int) {
	defer test.Truncate(t, store.db, orgId, "files")

	s3Bucket := "test-bucket"
	fileUUID := uuid.Must(uuid.NewUUID())
	initialFile := pgdb.FileParams{
		PackageId:  packageId,
		Name:       "test-file",
		FileType:   fileType.EDF,
		S3Bucket:   s3Bucket,
		S3Key:      "test/s3/key/a1b2.edf",
		ObjectType: objectType.Source,
		Size:       1024,
		CheckSum:   "",
		Sha256:     "",
		UUID:       fileUUID,
	}
	var actualInitialFileId string
	var actualInitialUpdatedAt time.Time
	actualInitialFiles, err := store.AddFiles(context.Background(), []pgdb.FileParams{initialFile})
	if assert.NoError(t, err) {
		assert.Len(t, actualInitialFiles, 1)
		actualInitialFile := actualInitialFiles[0]
		assert.Equal(t, s3Bucket, actualInitialFile.S3Bucket)
		assert.Equal(t, initialFile.S3Key, actualInitialFile.S3Key)
		assert.Equal(t, fileUUID, actualInitialFile.UUID)
		actualInitialFileId = actualInitialFile.Id
		assert.NotEmpty(t, actualInitialFileId)
		actualInitialUpdatedAt = actualInitialFile.UpdatedAt
	}

	mistakeFile := pgdb.FileParams{
		PackageId:  packageId,
		Name:       "test-file",
		FileType:   fileType.EDF,
		S3Bucket:   s3Bucket,
		S3Key:      "test/not/the/same/key/a1b2.edf",
		ObjectType: objectType.Source,
		Size:       1024,
		CheckSum:   "",
		Sha256:     "",
		UUID:       fileUUID,
	}
	actualMistakeFiles, err := store.AddFiles(context.Background(), []pgdb.FileParams{mistakeFile})
	if assert.NoError(t, err) {
		assert.Empty(t, actualMistakeFiles)
	}
	var actualFileCount int
	err = store.db.QueryRow("SELECT count(*) from files where package_id = $1", packageId).Scan(&actualFileCount)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, actualFileCount)
	}

	var actualPacakgeId int
	var actualId, actualBucket, actualKey string
	var actualUUID uuid.UUID
	var actualUpdatedAt time.Time
	err = store.db.QueryRow("SELECT id, package_id, s3_bucket, s3_key, uuid, updated_at from files where package_id = $1", packageId).Scan(
		&actualId,
		&actualPacakgeId,
		&actualBucket,
		&actualKey,
		&actualUUID,
		&actualUpdatedAt)
	if assert.NoError(t, err) {
		assert.Equal(t, actualInitialFileId, actualId)
		assert.Equal(t, actualPacakgeId, packageId)
		assert.Equal(t, actualBucket, s3Bucket)
		assert.Equal(t, actualKey, initialFile.S3Key)
		assert.Equal(t, fileUUID, actualUUID)
		assert.Equal(t, actualInitialUpdatedAt, actualUpdatedAt)
	}
}
