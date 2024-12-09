package user

import (
	"fmt"
)

// Claim combines the role of the user in the org, and the features in the organization.
type Claim struct {
	Id           int64
	NodeId       string
	IsSuperAdmin bool
}

func (c Claim) String() string {
	return fmt.Sprintf("{ Id: %d, NodeId: %s, IsSuperAdmin: %t }", c.Id, c.NodeId, c.IsSuperAdmin)
}
