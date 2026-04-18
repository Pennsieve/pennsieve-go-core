package manifestFile

type Status int64

// Upload pipeline status flows:
// Local --> Registered --> Uploaded --> Imported --> Finalized --> Verified
// Local --> Registered --> Removed -->  <Deleted>
// Local --> Registered --> Changed --> Synced --> Uploaded --> Imported ....
// Local --> Registered --> Uploaded --> Failed
// Local --> Registered --> FailedOrphan (reconciler: object missing in S3)

const (
	Local        Status = iota // set when client creates upload (local only)
	Registered                 // set when client syncs with server (local, server)
	Imported                   // set when uploader imported file (local, server)
	Finalized                  // set when importer moves file to final destination (local, server)
	Verified                   // set when client is informed of finalized status (local, server)
	Failed                     // set when importer fails to import (local, server)
	Removed                    // set when client removes file locally (local only)
	Unknown                    // set when sync failed (local only)
	Changed                    // set when synced file is updated locally (local only)
	Uploaded                   // set when file has successfully been uploaded to server (local only)
	FailedOrphan               // set by server-side reconciler when a Registered file's S3 object is confirmed missing past the grace period (server-originated terminal)
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
	case FailedOrphan:
		return "FailedOrphan"
	default:
		return "Initiated"
	}
}

// IsInProgress returns "x" when a file's status still expects more automatic
// activity (so it should appear in the sparse InProgressIndex and keep its
// manifest in a non-Completed state), and "" when the file has reached a
// terminal state that no background worker will change on its own.
//
// Terminal set:
//
//   - Imported, Finalized, Verified — successful import pipeline stages.
//   - Removed                       — explicitly dropped from the manifest.
//   - Failed                        — import hit an unrecoverable error;
//     no automatic retry mechanism drives it forward.
//   - FailedOrphan                  — server-side reconciler confirmed the
//     S3 object is missing; terminal unless the client explicitly re-syncs
//     (which is an operator-triggered state transition, not automatic).
//
// Non-terminal (in progress): Local, Registered, Uploaded, Changed, Unknown.
func (s Status) IsInProgress() string {
	switch s {
	case Imported, Verified, Finalized, Removed, Failed, FailedOrphan:
		return ""
	}
	return "x"
}

// ManifestFileStatusMap maps string values to FileStatus objects.
//
// Unknown strings map to Unknown (not Local) so callers don't misinterpret a
// future-version status as "file hasn't been synced yet" and kick off a
// spurious re-upload.
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
	case "FailedOrphan":
		return FailedOrphan
	}
	return Unknown
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
	UploadId string `json:"upload_id"`
	Status   string `json:"status"`
	Icon     string `json:"icon"`
}

type GetManifestFilesResponse struct {
	ManifestId        string `json:"manifest_id"`
	Files             []DTO  `json:"files"`
	ContinuationToken string `json:"continuation_token"` //Upload ID of the last returned item
}
