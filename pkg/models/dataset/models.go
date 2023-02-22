package dataset

import (
	"fmt"
	"strings"
)

type Role int64

const (
	None Role = iota
	Viewer
	Editor
	Manager
	Owner
)

var RoleMap = map[string]Role{
	"none":    None,
	"viewer":  Viewer,
	"editor":  Editor,
	"manager": Manager,
	"owner":   Owner,
}

func RoleFromString(str string) (Role, bool) {
	c, ok := RoleMap[strings.ToLower(str)]
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

// Claim provides an object that describes a Role and a Target
type Claim struct {
	Role   Role
	NodeId string
	IntId  int64
}

func (c Claim) String() string {
	return fmt.Sprintf("%s (%d) - %s", c.NodeId, c.IntId, c.Role.String())
}
