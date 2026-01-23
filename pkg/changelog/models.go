package changelog

import (
	"fmt"
	"time"
)

type Type int64

// TODO: add other types based on API
const (
	CreatePackage Type = iota
	DeletePackage
	RestorePackage
)

const (
	createPackageString  = "CREATE_PACKAGE"
	deletePackageString  = "DELETE_PACKAGE"
	restorePackageString = "RESTORE_PACKAGE"
	unknownTypeString    = "UNKNOWN"
)

func (s Type) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Type) UnmarshalText(text []byte) error {
	switch string(text) {
	case createPackageString:
		*s = CreatePackage
	case deletePackageString:
		*s = DeletePackage
	case restorePackageString:
		*s = RestorePackage
	default:
		return fmt.Errorf("unknown changelog type: %s", text)
	}
	return nil
}

func (s Type) String() string {
	switch s {
	case CreatePackage:
		return createPackageString
	case DeletePackage:
		return deletePackageString
	case RestorePackage:
		return restorePackageString
	}

	return unknownTypeString
}

type MessageParams struct {
	OrganizationId int64   `json:"organizationId"`
	DatasetId      int64   `json:"datasetId"`
	UserId         string  `json:"userId"`
	Events         []Event `json:"events"`
	TraceId        string  `json:"traceId"`
	Id             string  `json:"id"`
}

type ParentPackage struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	NodeId string `json:"nodeId"`
}

type PackageCreateEvent struct {
	Id     int64          `json:"id"`
	Name   string         `json:"name"`
	NodeId string         `json:"nodeId"`
	Parent *ParentPackage `json:"parent"`
}

type PackageRestoreEvent struct {
	Id           int64          `json:"id"`
	Name         string         `json:"name,omitempty"`
	OriginalName string         `json:"originalName,omitempty"`
	NodeId       string         `json:"nodeId,omitempty"`
	Parent       *ParentPackage `json:"parent,omitempty"`
}

type Event struct {
	EventType   Type        `json:"eventType"`
	EventDetail interface{} `json:"eventDetail"`
	Timestamp   time.Time   `json:"timestamp"`
}

type Message struct {
	DatasetChangelogEventJob MessageParams `json:"DatasetChangelogEventJob"`
}
