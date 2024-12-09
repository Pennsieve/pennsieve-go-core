package organization

import (
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
)

// Claim combines the role of the user in the org, and the features in the organization.
type Claim struct {
	Role            pgdb.DbPermission
	IntId           int64
	NodeId          string
	EnabledFeatures []pgdb.FeatureFlags
}

func (c Claim) String() string {
	return fmt.Sprintf("{ Id: %d, NodeId: %s, Role: %d (%s), EnabledFeatures: %+v }",
		c.IntId, c.NodeId, c.Role, c.Role.String(), c.EnabledFeatures)
}

// HasRole returns true if this claim contains permissions sufficient to satisfy the given requiredOrgRole
func (c Claim) HasRole(requiredOrgRole role.Role) bool {
	return c.Role.ImpliesRole(requiredOrgRole)
}
