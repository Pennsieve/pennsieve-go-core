package pgdb

import (
	"time"
)

type DbPermission int64

const (
	NoPermission DbPermission = 0
	Collaborate               = 1
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
	case Collaborate:
		return "Collaborate"
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

type OrganizationUser struct {
	OrganizationId int64        `json:"organization_id"`
	UserId       int64        `json:"user_id"`
	DbPermission DbPermission `json:"permission_bit"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}
