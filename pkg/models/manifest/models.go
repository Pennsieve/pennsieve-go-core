package manifest

import (
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
)

type Status int64

const (
	Initiated Status = iota
	Uploading
	Completed
	Cancelled
	Archived
)

func (s Status) String() string {
	switch s {
	case Initiated:
		return "Initiated"
	case Uploading:
		return "Uploading"
	case Completed:
		return "Completed"
	case Cancelled:
		return "Cancelled"
	case Archived:
		return "Archived"
	default:
		return "Initiated"
	}
}

func (s Status) ManifestStatusMap(value string) Status {
	switch value {
	case "Initiated":
		return Initiated
	case "Uploading":
		return Uploading
	case "Completed":
		return Completed
	case "Cancelled":
		return Cancelled
	case "Archived":
		return Archived
	}
	return Initiated
}

type DTO struct {
	ID        string                 `json:"id"`
	DatasetId string                 `json:"dataset_id"`
	Files     []manifestFile.FileDTO `json:"files"`
	Status    Status                 `json:"status"`
}

type ManifestDTO struct {
	Id            string `json:"id"`
	DatasetId     int64  `json:"dataset_id"`
	DatasetNodeId string `json:"dataset_node_id"`
	Status        string `json:"status"`
	User          int64  `json:"user"`
	DateCreated   int64  `json:"date_created"`
}

type GetResponse struct {
	Manifests []ManifestDTO `json:"manifests"`
}

type GetStatusEndpointResponse struct {
	ManifestId        string   `json:"manifest_id"`
	Status            string   `json:"status"`
	Files             []string `json:"files"`
	ContinuationToken string   `json:"continuation_token"`
	Verified          bool     `json:"verified"`
}

type PostResponse struct {
	ManifestNodeId string                       `json:"manifest_node_id"`
	NrFilesUpdated int                          `json:"nr_files_updated"`
	NrFilesRemoved int                          `json:"nr_files_removed"`
	UpdatedFiles   []manifestFile.FileStatusDTO `json:"updated_files"`
	FailedFiles    []string                     `json:"failed_files"`
}

// AddFilesStats object that is returned to the client.
type AddFilesStats struct {
	NrFilesUpdated int
	NrFilesRemoved int
	FileStatus     []manifestFile.FileStatusDTO
	FailedFiles    []string
}
