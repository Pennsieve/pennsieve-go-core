package organization

import (
	"fmt"
	dbTable2 "github.com/pennsieve/pennsieve-go-core/pkg/models/dbTable"
)

// Claim combines the role of the user in the org, and the features in the organization.
type Claim struct {
	Role            dbTable2.DbPermission
	IntId           int64
	EnabledFeatures []dbTable2.FeatureFlags
}

func (c Claim) String() string {
	return fmt.Sprintf("OrganizationId: %d - %s", c.IntId, c.Role.String())
}
