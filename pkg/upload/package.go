package upload

import (
	"strings"

	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
	log "github.com/sirupsen/logrus"
)

func PackageTypeResolver(items []manifestFile.FileDTO) []manifestFile.FileDTO {

	for i := range items {

		// Only resolve the type from the file name when one hasn't already
		// been set on the FileDTO; an existing FileType is left untouched.
		if len(items[i].FileType) == 0 {
			fType := resolveTypeFromName(items[i].TargetName)
			items[i].FileType = fType.String()
		}
	}
	return items
}

// resolveTypeFromName determines a file's fileType.Type from its name by
// matching against the longest known extension in ExtensionToTypeDict.
// It uses a longest-suffix match rather than splitting on the first '.' so that
// names with extra dots (e.g. "example.june_21.edf") and multi-part extensions
// (e.g. "scan.nii.gz", "image.ome.tiff") both resolve correctly. Case-insensitive.
func resolveTypeFromName(targetName string) fileType.Type {
	lowerTargetName := strings.ToLower(targetName)

	var bestKey string
	bestLen := -1
	for ext := range fileType.ExtensionToTypeDict {
		// Most keys are stored without a leading dot ("edf", "nii.gz")
		// But a few keep the leading `.` (e.g. ".eeg")
		// Normalize by trimming any leading dots.
		trimmedExt := strings.TrimPrefix(ext, ".")
		if trimmedExt == "" {
			continue
		}
		// longer matches are always the right choice (ie choose ".nii.gz" over ".gz")
		if len(trimmedExt) > bestLen && strings.HasSuffix(lowerTargetName, "."+trimmedExt) {
			bestLen = len(trimmedExt)
			bestKey = ext
		}
	}

	if bestKey == "" {
		log.WithFields(log.Fields{"targetName": targetName}).
			Debug("no known file extension matched; defaulting to GenericData")
		return fileType.GenericData
	}
	return fileType.ExtensionToTypeDict[bestKey]
}

func persystMerger(fileName string, layFile *manifestFile.FileDTO, items []manifestFile.FileDTO) {

	// Iterate over files and if file exists in same folder with same name and ".dat" extension, than merge the two.
	// Then set MergePackageID for both lay and dat file to the uploadID of the lay file.
	// This ensures that when we create the package in upload_handler that we set the name of the package to be
	// the filename without extension.
	for i, f := range items {
		if layFile.TargetPath == f.TargetPath && layFile.TargetName != f.TargetName {
			if strings.HasPrefix(f.TargetName, fileName) && strings.HasSuffix(f.TargetName, ".dat") {
				items[i].MergePackageId = layFile.UploadID
				layFile.MergePackageId = layFile.UploadID
				items[i].FileType = fileType.Persyst.String()
				log.Debug("Found match in: ", f.TargetName)
				break
			}
		}
	}
}
