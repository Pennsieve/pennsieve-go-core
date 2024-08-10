package pgdb

import (
	"strings"
	"time"
)

type DbPermission int64

const (
	NoPermission DbPermission = 0
	Guest                     = 1
	Read                      = 2
	Write                     = 4
	Delete                    = 8
	Administer                = 16
	Owner                     = 32
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

func (s DbPermission) AsRoleString() string {
	switch s {
	case NoPermission:
		return "none"
	case Guest:
		return "guest"
	case Read:

		return "viewer"
	case Write:

		return "editor"
	case Delete:

		return "editor"
	case Administer:

		return "manager"
	case Owner:

		return "owner"
	default:
		return "none"
	}
}

type OrganizationUser struct {
	OrganizationId int64        `json:"organization_id"`
	UserId         int64        `json:"user_id"`
	DbPermission   DbPermission `json:"permission_bit"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}
