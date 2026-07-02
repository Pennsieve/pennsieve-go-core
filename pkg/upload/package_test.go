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

// TestPackageTypeResolverExtensions exercises the longest-suffix extension
// matching, with emphasis on the cases the old first-dot split got wrong:
// names with extra dots before the extension, and multi-part extensions.
func TestPackageTypeResolverExtensions(t *testing.T) {
	tests := []struct {
		name         string
		targetName   string
		expectedType string
	}{
		// Single-part extensions
		{"single dot", "example.edf", "EDF"},
		{"extra dots before single-part ext (the bug)", "example.june_21.edf", "EDF"},
		{"many extra dots", "a.b.c.d.edf", "EDF"},

		// Multi-part extensions
		{"multi-part nii.gz", "scan.nii.gz", "NIFTI"},
		{"multi-part with extra dots", "subject.2024.session1.nii.gz", "NIFTI"},
		{"multi-part ome.tiff", "image.ome.tiff", "OMETIFF"},

		// Longest-suffix precedence: nii.gz/tar.gz must win over bare gz
		{"longest match nii.gz beats gz", "brain.nii.gz", "NIFTI"},
		{"longest match tar.gz beats gz", "archive.tar.gz", "ZIP"},
		{"bare gz still resolves", "data.gz", "ZIP"},

		// Leading-dot dict key (".eeg") still matches via normalization
		{"leading-dot dict key eeg", "recording.eeg", "NihonKoden"},

		// Case-insensitivity
		{"uppercase single", "EXAMPLE.EDF", "EDF"},
		{"mixed-case multi-part", "Scan.NII.GZ", "NIFTI"},

		// Non-matches default to GenericData
		{"unknown extension", "notes.xyz", "GenericData"},
		{"no extension", "README", "GenericData"},
		{"dotfile, no known ext", ".bashrc", "GenericData"},
		{"name equals ext but no dot", "edf", "GenericData"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := []manifestFile.FileDTO{{
				UploadID:   "0",
				TargetPath: "path",
				TargetName: tt.targetName,
				FileType:   "",
			}}
			result := PackageTypeResolver(files)
			assert.Equal(t, tt.expectedType, result[0].FileType,
				"file %q should resolve to %s", tt.targetName, tt.expectedType)
		})
	}
}

// TestPackageTypeResolverPreservesExistingType ensures a FileType that is
// already set is never overwritten, even when the name would resolve elsewhere.
func TestPackageTypeResolverPreservesExistingType(t *testing.T) {
	files := []manifestFile.FileDTO{{
		UploadID:   "0",
		TargetPath: "path",
		TargetName: "example.june_21.edf",
		FileType:   "DICOM",
	}}
	result := PackageTypeResolver(files)
	assert.Equal(t, "DICOM", result[0].FileType,
		"an already-set FileType must not be overwritten")
}
