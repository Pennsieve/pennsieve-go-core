package manifest

import (
	"github.com/pennsieve/pennsieve-go-core/models/manifest/manifestFile"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestPackage is main testing function that sets up tests and runs sub-tests.
func TestPackage(t *testing.T) {
	// <setup code>
	s := ManifestSession{
		FileTableName: "",
		TableName:     "",
		Client:        nil,
		SNSClient:     nil,
		SNSTopic:      "",
	}

	files := []manifestFile.FileDTO{
		{
			UploadID:       "0",
			TargetPath:     "path1",
			TargetName:     "file1.txt",
			MergePackageId: "",
			FileType:       "",
		}, {
			UploadID:       "1",
			TargetPath:     "path1",
			TargetName:     "file1.fastq.gz",
			MergePackageId: "",
			FileType:       "",
		}, {
			UploadID:       "2",
			TargetPath:     "path1",
			TargetName:     "file1.gz",
			MergePackageId: "",
			FileType:       "",
		}, {
			UploadID:       "3",
			TargetPath:     "path1",
			TargetName:     "persyst.dat",
			MergePackageId: "",
			FileType:       "",
		},
		{
			UploadID:       "4",
			TargetPath:     "path1",
			TargetName:     "persyst.lay",
			MergePackageId: "",
			FileType:       "",
		},
		{
			UploadID:       "5",
			TargetPath:     "path1",
			TargetName:     "persyst.unknown",
			MergePackageId: "",
			FileType:       "",
		},
		{
			UploadID:       "6",
			TargetPath:     "path2",
			TargetName:     "persyst2.lay",
			MergePackageId: "",
			FileType:       "",
		},
		{
			UploadID:       "7",
			TargetPath:     "path1",
			TargetName:     "persyst2.dat",
			MergePackageId: "",
			FileType:       "",
		},
	}
	processedFiles := s.PackageTypeResolver(files)

	t.Run("BasicExtensions", func(t *testing.T) {
		testBasicExtensions(t, processedFiles)
	})
	t.Run("Persyst", func(t *testing.T) {
		testPersystMerging(t, processedFiles)
	})
	// <tear-down code>
}

func testBasicExtensions(t *testing.T, files []manifestFile.FileDTO) {
	assert.Equal(t, "Text", files[0].FileType,
		"Extension '.txt' should return Text type.")

	assert.Equal(t, "FASTQ", files[1].FileType,
		"Extension '.fastq.gz' should return FASTQ type.")

	assert.Equal(t, "ZIP", files[2].FileType,
		"Extension '.gz' should return ZIP type")

	assert.Equal(t, "GenericData", files[5].FileType,
		"Unknown extensions should return 'Generic Data' type.")
}

func testPersystMerging(t *testing.T, files []manifestFile.FileDTO) {
	// Check Persyst Merging and file-type
	assert.Equal(t, "Persyst", files[3].FileType,
		"Extension '.dat' should return PERSYST type if a '.lay' file with same name exists in folder.")
	assert.Equal(t, "Persyst", files[4].FileType,
		"Extesnsion '.lay' should return PERSYST type")

	assert.Equal(t, "4", files[3].MergePackageId,
		"'.dat' file should have merge-id belonging to corresponding '.lay' file.")
	assert.Equal(t, "4", files[4].MergePackageId,
		"'.lay' file should have merge-id point to itself if there is a corresponding '.dat' file.")

	assert.Emptyf(t, files[6].MergePackageId,
		"'.Lay' file with '.dat' file in different directory should not be merged.")
	assert.Equal(t, "Data", files[7].FileType)

}
