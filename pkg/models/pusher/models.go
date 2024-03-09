package pusher

import "database/sql"

type Config struct {
	AppId   string `json:"app_id"`
	Key     string `json:"key"`
	Secret  string `json:"secret"`
	Cluster string `json:"cluster"`
}

type UploadMessageItem struct {
	Name     string         `json:"name"`
	NodeId   string         `json:"node_id"`
	ParentId sql.NullInt64  `json:"parent_id,omitempty"`
	UploadId sql.NullString `json:"upload_id,omitempty"`
}
