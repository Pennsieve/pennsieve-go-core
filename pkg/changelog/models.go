package changelog

type Type int64

// TODO: add other types based on API
const (
	createPackage Type = iota
	deletePackage
)

func (s Type) String() string {
	switch s {
	case createPackage:
		return "CREATE_PACKAGE"
	case deletePackage:
		return "DELETE_PACKAGE"
	}

	return "UNKNOWN"
}

type MessageParams struct {
	OrganizationId int64         `json:"OrganizationId"`
	DatasetId      int64         `json:"datasetId"`
	UserId         int64         `json:"userId"`
	Events         []interface{} `json:"events"`
	TraceId        string        `json:"traceId"`
	Id             string        `json:"id"`
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
