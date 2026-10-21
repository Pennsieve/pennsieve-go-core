package nodeId

import (
	"fmt"
	"github.com/google/uuid"
)

type NodeType int64

const (
	UserCode NodeType = iota
	OrganizationCode
	TeamCode
	FileCode
	PackageCode
	CollectionCode
	DataSetCode
	ChannelCode
	_ // was DataCanvasCode; kept so FolderCode keeps its value
	FolderCode
)

var NodeCodeMap = map[NodeType]string{
	UserCode:         "user",
	OrganizationCode: "organization",
	TeamCode:         "team",
	FileCode:         "file",
	PackageCode:      "package",
	CollectionCode:   "collection",
	DataSetCode:      "dataset",
	ChannelCode:      "channel",
	FolderCode:       "folder",
}

func NodeId(nodeType NodeType) string {
	return fmt.Sprintf("N:%s:%s", NodeCodeMap[nodeType], uuid.NewString())
}
