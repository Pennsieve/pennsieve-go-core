package role

import (
	"strings"
)

type Role int64

const (
	None Role = iota
	Guest
	Viewer
	Editor
	Manager
	Owner
)

var Map = map[string]Role{
	"none":    None,
	"guest":   Guest,
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
	case Guest:
		return "Guest"
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

// Implies returns true if this Role implies the given requiredRole and false otherwise.
// That is, if a user has this Role and an action requires requiredRole, this method
// returns true if the user can perform the operation and false if they cannot.
func (s Role) Implies(requiredRole Role) bool {
	return s >= requiredRole
}
