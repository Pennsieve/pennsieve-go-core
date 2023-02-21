package objectType

type ObjectType int64

const (
	View ObjectType = iota
	File
	Source
)

func (p ObjectType) String() string {
	switch p {
	case View:
		return "view"
	case File:
		return "file"
	case Source:
		return "source"
	default:
		return "file"
	}
}

var Dict = map[string]ObjectType{
	"view":   View,
	"file":   File,
	"source": Source,
}
