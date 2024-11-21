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
	return fmt.Sprintf("%s (%d) - %s", c.NodeId, c.IntId, c.Role.String())
}
