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
	return fmt.Sprintf("Name: %s (id: %d nodeId: %s permission: %d)",
		c.Name, c.IntId, c.NodeId, c.Permission)
}
