package datasetType

import "strings"

type DatasetType int

const (
	Research DatasetType = iota + 1
	Trial
	Collection
	Release
)

var Map = map[string]DatasetType{
	"research":   Research,
	"trial":      Trial,
	"collection": Collection,
	"release":    Release,
}

func DatasetTypeFromString(str string) (DatasetType, bool) {
	c, ok := Map[strings.ToLower(str)]
	return c, ok
}

func (s DatasetType) String() string {
	switch s {
	case Research:
		return "research"
	case Trial:
		return "trial"
	case Collection:
		return "collection"
	case Release:
		return "release"
	default:
		return "research"
	}
}
