package organization

import (
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/models/dbTable"
)

// Claim combines the role of the user in the org, and the features in the organization.
type Claim struct {
	Role            dbTable.DbPermission
	IntId           int64
	EnabledFeatures []dbTable.FeatureFlags
}

func (c Claim) String() string {
	return fmt.Sprintf("OrganizationId: %d - %s", c.IntId, c.Role.String())
}
