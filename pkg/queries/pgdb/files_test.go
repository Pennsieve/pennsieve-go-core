package pgdb

import (
	"context"
	"encoding/json"
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
		"AddFiles empty slice is a no-op":            testAddFilesEmptySlice,
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

// testAddFilesEmptySlice is the regression test for
// `pq: syntax error at or near "ON"`.
//
// Before the guard in AddFiles, an empty batch produced
// "INSERT INTO files(...) VALUES  ON CONFLICT ..." — an empty VALUES list — and
// PrepareContext failed with that syntax error. It surfaced in production when a
// warm Lambda's dedup map filtered a redelivered SQS batch down to zero files;
// the error named the SQL rather than the empty input, which sent the
// investigation in the wrong direction.
func testAddFilesEmptySlice(t *testing.T, store *SQLStore, orgId int, packageId int) {
	defer test.Truncate(t, store.db, orgId, "files")

	for _, empty := range [][]pgdb.FileParams{nil, {}} {
		actual, err := store.AddFiles(context.Background(), empty)
		if assert.NoError(t, err, "empty batch must not error") {
			assert.Empty(t, actual, "empty batch must insert nothing")
		}
	}
}

// TestBuildAddFilesQuery covers the generated SQL without needing a database.
func TestBuildAddFilesQuery(t *testing.T) {
	now := time.Date(2026, 7, 13, 2, 24, 31, 0, time.UTC)

	newParams := func(checksum, sha string) pgdb.FileParams {
		return pgdb.FileParams{
			PackageId:  1,
			Name:       "f.edf",
			FileType:   fileType.EDF,
			S3Bucket:   "b",
			S3Key:      "k",
			ObjectType: objectType.Source,
			Size:       1,
			CheckSum:   checksum,
			Sha256:     sha,
			UUID:       uuid.Must(uuid.NewUUID()),
		}
	}

	t.Run("single row placeholders and ON CONFLICT spacing", func(t *testing.T) {
		sql, values := buildAddFilesQuery([]pgdb.FileParams{newParams("", "")}, now)

		assert.Contains(t, sql, "VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) ON CONFLICT")
		// The bug this guards: ")ON CONFLICT" parses, but only by luck.
		assert.NotContains(t, sql, ")ON CONFLICT")
		assert.Len(t, values, filesPerRowColumns)
	})

	t.Run("second row continues placeholder numbering", func(t *testing.T) {
		sql, values := buildAddFilesQuery(
			[]pgdb.FileParams{newParams("", ""), newParams("", "")}, now)

		assert.Contains(t, sql, "($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13),"+
			"($14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26) ON CONFLICT")
		assert.Len(t, values, 2*filesPerRowColumns)
	})

	// Pins the exact bytes of the checksum JSON for values needing no escaping,
	// so this refactor cannot change what is already stored in the column.
	t.Run("checksum JSON is byte-identical for safe values", func(t *testing.T) {
		_, values := buildAddFilesQuery([]pgdb.FileParams{newParams("abc123", "ZGVmNDU2")}, now)

		assert.Equal(t,
			`{"checksum": "abc123", "chunkSize": "32", "sha256": "ZGVmNDU2"}`,
			values[7],
		)
	})

	t.Run("checksum JSON stays valid when values need escaping", func(t *testing.T) {
		_, values := buildAddFilesQuery(
			[]pgdb.FileParams{newParams(`he"llo`, `back\slash`)}, now)

		etag, ok := values[7].(string)
		if !assert.True(t, ok, "etag should be a string") {
			return
		}

		var decoded map[string]string
		if assert.NoError(t, json.Unmarshal([]byte(etag), &decoded),
			"etag must be valid JSON even when values contain quotes or backslashes") {
			assert.Equal(t, `he"llo`, decoded["checksum"])
			assert.Equal(t, `back\slash`, decoded["sha256"])
			assert.Equal(t, "32", decoded["chunkSize"])
		}
	})
}
