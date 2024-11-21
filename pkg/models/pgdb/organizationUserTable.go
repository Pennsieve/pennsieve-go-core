package pgdb

import (
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"strings"
	"time"
)

type DbPermission int64

const (
	NoPermission DbPermission = 0
	Guest        DbPermission = 1
	Read         DbPermission = 2
	Write        DbPermission = 4
	Delete       DbPermission = 8
	Administer   DbPermission = 16
	Owner        DbPermission = 32
)

func (s DbPermission) String() string {
	switch s {
	case NoPermission:
		return "NoPermission"
	case Guest:
		return "Guest"
	case Read:
		return "Read"
	case Write:
		return "Write"
	case Delete:
		return "Delete"
	case Administer:
		return "Administer"
	case Owner:
		return "Owner"
	}

	return "NoPermission"
}

func FromRole(role string) DbPermission {
	switch strings.ToLower(role) {
	case "guest":
		return Guest
	case "viewer":
		return Read
	case "editor":
		return Delete
	case "manager":
		return Administer
	case "owner":
		return Owner
	default:
		return NoPermission
	}
}

func (s DbPermission) ToRole() role.Role {
	switch s {
	case NoPermission:
		return role.None
	case Guest:
		return role.Guest
	case Read:
		return role.Viewer
	case Write, Delete:
		return role.Editor
	case Administer:
		return role.Manager
	case Owner:
		return role.Owner
	default:
		return role.None
	}
}

// ImpliesRole returns true if this DbPermission implies the given requiredRole and false otherwise
// That is, if a user has this DbPermission and an action requires requiredRole, then the
// user is authorized to perform the action if this method returns true
func (s DbPermission) ImpliesRole(requiredRole role.Role) bool {
	return s.ToRole().Implies(requiredRole)
}

func (s DbPermission) AsRoleString() string {
	return strings.ToLower(s.ToRole().String())
}

type OrganizationUser struct {
	OrganizationId int64        `json:"organization_id"`
	UserId         int64        `json:"user_id"`
	DbPermission   DbPermission `json:"permission_bit"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}
