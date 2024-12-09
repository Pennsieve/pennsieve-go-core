package dataset

import (
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
)

// Claim provides an object that describes a Role and a Target
type Claim struct {
	Role   role.Role
	NodeId string
	IntId  int64
}

func (c Claim) String() string {
	return fmt.Sprintf("{ Id: %d, NodeId: %s, Role: %d (%s) }", c.IntId, c.NodeId, c.Role, c.Role.String())
}
