package upload

import (
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestPackage is main testing function that sets up tests and runs sub-tests.
func TestPackage(t *testing.T) {
	// <setup code>

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
	processedFiles := PackageTypeResolver(files)

	t.Run("BasicExtensions", func(t *testing.T) {
		testBasicExtensions(t, processedFiles)
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
