package manifestFile

type Status int64

// Upload pipeline status flows:
// Local --> Registered --> Uploaded --> Imported --> Finalized --> Verified
// Local --> Registered --> Removed -->  <Deleted>
// Local --> Registered --> Changed --> Synced --> Uploaded --> Imported ....
// Local --> Registered --> Uploaded --> Failed

const (
	Local      Status = iota // set when client creates manifest (local only)
	Registered               // set when client syncs with server (local, server)
	Imported                 // set when uploader imported file (local, server)
	Finalized                // set when importer moves file to final destination (local, server)
	Verified                 // set when client is informed of finalized status (local, server)
	Failed                   // set when importer fails to import (local, server)
	Removed                  // set when client removes file locally (local only)
	Unknown                  // set when sync failed (local only)
	Changed                  // set when synced file is updated locally (local only)
	Uploaded                 // set when file has successfully been uploaded to server (local only)
)

// String returns string version of FileStatus object.
func (s Status) String() string {
	switch s {
	case Local:
		return "Local"
	case Registered:
		return "Registered"
	case Imported:
		return "Imported"
	case Finalized:
		return "Finalized"
	case Verified:
		return "Verified"
	case Failed:
		return "Failed"
	case Removed:
		return "Removed"
	case Unknown:
		return "Unknown"
	case Changed:
		return "Changed"
	case Uploaded:
		return "Uploaded"
	default:
		return "Initiated"
	}
}

// IsInProgress returns a boolean indicating whether upload status reflects a finalized status
func (s Status) IsInProgress() string {
	if s == Imported || s == Verified || s == Finalized || s == Removed {
		return ""
	}
	return "x"
}

// ManifestFileStatusMap maps string values to FileStatus objects.
func (s Status) ManifestFileStatusMap(value string) Status {
	switch value {
	case "Local":
		return Local
	case "Registered":
		return Registered
	case "Imported":
		return Imported
	case "Finalized":
		return Finalized
	case "Verified":
		return Verified
	case "Removed":
		return Removed
	case "Failed":
		return Failed
	case "Changed":
		return Changed
	case "Uploaded":
		return Uploaded
	case "Unknown":
		return Unknown
	}
	return Local
}

// FileDTO used to transfer file object in API requests.
type FileDTO struct {
	UploadID       string `json:"upload_id"`
	S3Key          string `json:"s3_key"`
	TargetPath     string `json:"target_path"`
	TargetName     string `json:"target_name"`
	Status         Status `json:"status"`
	MergePackageId string `json:"merge_package_id"` // MergePackageId is ID of package if not using UploadID
	FileType       string `json:"file_type"`        // FileType is string representation of fileType (auto populated if empty)
}

// FileStatusDTO used to transfer status information in API requests.
type FileStatusDTO struct {
	UploadId string `json:"upload_id"`
	Status   Status `json:"status"`
}

type DTO struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	FileType string `json:"file_type"`
	UploadId string `json:"upload_id""`
	Status   string `json:"status"`
	Icon     string `json:"icon"`
}

type GetManifestFilesResponse struct {
	ManifestId        string `json:"manifest_id"`
	Files             []DTO  `json:"files"`
	ContinuationToken string `json:"continuation_token"` //Upload ID of the last returned item
}
