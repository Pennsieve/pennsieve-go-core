package uploadFile

import (
	"encoding/json"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/iconInfo"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageType"
	"sort"
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
