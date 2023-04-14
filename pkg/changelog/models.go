package changelog

import "time"

type Type int64

// TODO: add other types based on API
const (
	CreatePackage Type = iota
	DeletePackage
)

func (s Type) String() string {
	switch s {
	case CreatePackage:
		return "CREATE_PACKAGE"
	case DeletePackage:
		return "DELETE_PACKAGE"
	}

	return "UNKNOWN"
}

type MessageParams struct {
	OrganizationId int64   `json:"organizationId"`
	DatasetId      int64   `json:"datasetId"`
	UserId         int64   `json:"userId"`
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
	parent *ParentPackage `json:"parent"`
}

type Event struct {
	EventType   Type        `json:"eventType"`
	EventDetail interface{} `json:"eventDetail"`
	Timestamp   time.Time   `json:"timestamp"`
}

type Message struct {
	DatasetChangelogEventJob MessageParams `json:"DatasetChangelogEventJob"`
}
