package uploadFile

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/iconInfo"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/uploadFolder"
	"regexp"
	"sort"
	"strings"
)

// UploadFile is the parsed and cleaned representation of the SQS S3 Put Event
type UploadFile struct {
	ManifestId     string           // ManifestId is id for the entire upload session.
	UploadId       string           // UploadId ID is used as part of the s3key for uploaded files and is packageID.
	S3Bucket       string           // S3Bucket is bucket where file is uploaded to
	S3Key          string           // S3Key is the S3 key of the file
	Path           string           // Path to collection without file-name
	Name           string           // Name is the filename including extension(s)
	Extension      string           // Extension of file (separated from name)
	FileType       fileType.Type    // FileType is the type of the file
	Type           packageType.Type // Type of the Package.
	SubType        string           // SubType of the file
	Icon           iconInfo.Icon    // Icon for the file
	Size           int64            // Size of file
	ETag           string           // ETag provided by S3
	MergePackageId string           // MergePackageId is packageID leveraged instead of upload id in case of package merging
	Sha256         string           // Sha256 checksum of the file
}

// String returns a json representation of the UploadFile object
func (f *UploadFile) String() string {
	str, _ := json.Marshal(f)
	return string(str)
}

// Sort sorts []UploadFiles by the depth of the folder the file resides in.
func (f *UploadFile) Sort(files []UploadFile) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}

// GetUploadFolderMap returns an object that maps path name to Folder object.
func (f *UploadFile) GetUploadFolderMap(sortedFiles []UploadFile, targetFolder string) uploadFolder.UploadFolderMap {

	// Mapping path from targetFolder to UploadFolder Object
	var folderNameMap = map[string]*uploadFolder.UploadFolder{}

	// Iterate over the files and create the UploadFolder objects.
	for _, f := range sortedFiles {

		if f.Path == "" {
			continue
		}

		// Prepend the target-Folder if it exists
		p := f.Path
		if targetFolder != "" {
			p = strings.Join(
				[]string{
					targetFolder, p,
				}, "/")
		}

		// Remove leading and trailing "/"
		leadingSlashes := regexp.MustCompile(`^\/+`)
		p = leadingSlashes.ReplaceAllString(p, "")

		trailingSlashes := regexp.MustCompile(`\/+$`)
		p = trailingSlashes.ReplaceAllString(p, "")

		// Iterate over path segments in a single file and identify folders.
		pathSegments := strings.Split(p, "/")
		absoluteSegment := "" // Current location in the path walker for current file.
		currentNodeId := ""
		currentFolderPath := ""
		for depth, segment := range pathSegments {

			parentNodeId := currentNodeId
			parentFolderPath := currentFolderPath

			// If depth > 0 ==> prepend the previous absoluteSegment to the current path name.
			if depth > 0 {
				absoluteSegment = strings.Join(
					[]string{

						absoluteSegment,
						segment,
					}, "/")
			} else {
				absoluteSegment = segment
			}

			// If folder already exists in map, add current folder as a child to the parent
			// folder (which will exist too at this point). If not, create new folder to the map and add to parent folder.

			folder, ok := folderNameMap[absoluteSegment]
			if ok {
				currentNodeId = folder.NodeId
				currentFolderPath = absoluteSegment

			} else {
				currentNodeId = fmt.Sprintf("N:collection:%s", uuid.New().String())
				currentFolderPath = absoluteSegment

				folder = &uploadFolder.UploadFolder{
					NodeId:       currentNodeId,
					Name:         segment,
					ParentNodeId: parentNodeId,
					ParentId:     -1,
					Depth:        depth,
				}
				folderNameMap[absoluteSegment] = folder
			}

			// Add current segment to parent if exist
			if parentFolderPath != "" {
				folderNameMap[parentFolderPath].Children = append(folderNameMap[parentFolderPath].Children, folder)
			}

		}
	}

	return folderNameMap
}
