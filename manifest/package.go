package manifest

import (
	"github.com/pennsieve/pennsieve-go-core/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/models/manifest/manifestFile"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

func (s ManifestSession) PackageTypeResolver(items []manifestFile.FileDTO) []manifestFile.FileDTO {

	for i, f := range items {

		// Determine Type based on extension, or
		// return type that is already defined in FileDTO
		var fileName, fileExtension string
		var fType fileType.Type
		if len(items[i].FileType) == 0 {
			// 1. Find FileType

			// Split on the first '.' and consider everything after the extension.
			r := regexp.MustCompile(`(?P<FileName>[^\.]*)?\.?(?P<Extension>.*)`)
			pathParts := r.FindStringSubmatch(f.TargetName)
			if pathParts == nil {
				log.WithFields(
					log.Fields{
						"upload_id": items[i].UploadID,
					},
				).Error("Unable to parse filename:", f.TargetName)
				continue
			}

			fileName = pathParts[r.SubexpIndex("FileName")]
			fileExtension = pathParts[r.SubexpIndex("Extension")]

			var exists bool
			fType, exists = fileType.ExtensionToTypeDict[fileExtension]
			if !exists {
				fType = fileType.GenericData
			}

			// Set the type if not previously set.
			items[i].FileType = fType.String()
		} else {
			fType = fileType.Dict[items[i].FileType]
		}

		// 2. Implement Merge Strategy
		switch fType {
		case fileType.Persyst:
			persystMerger(fileName, &items[i], items)
		default:
			continue

		}
	}
	return items
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
