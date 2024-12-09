package teamUser

import (
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
)

type Claim struct {
	IntId      int64
	Name       string
	NodeId     string
	Permission pgdb.DbPermission
	TeamType   string
}

func (c Claim) String() string {
	return fmt.Sprintf("{ Name: %s, Id: %d, NodeId: %s, Permission: %d (%s) }",
		c.Name, c.IntId, c.NodeId, c.Permission, c.Permission.String())
}
