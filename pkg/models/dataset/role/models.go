package role

import "strings"

type Role int64

const (
	None Role = iota
	Viewer
	Editor
	Manager
	Owner
)

var Map = map[string]Role{
	"none":    None,
	"viewer":  Viewer,
	"editor":  Editor,
	"manager": Manager,
	"owner":   Owner,
}

func RoleFromString(str string) (Role, bool) {
	c, ok := Map[strings.ToLower(str)]
	return c, ok
}

func (s Role) String() string {
	switch s {
	case None:
		return "None"
	case Viewer:
		return "Viewer"
	case Editor:
		return "Editor"
	case Manager:
		return "Manager"
	case Owner:
		return "Owner"
	}

	return "Viewer"
}
