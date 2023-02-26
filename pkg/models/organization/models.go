package organization

import (
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
)

// Claim combines the role of the user in the org, and the features in the organization.
type Claim struct {
	Role            pgdb.DbPermission
	IntId           int64
	EnabledFeatures []pgdb.FeatureFlags
}

func (c Claim) String() string {
	return fmt.Sprintf("OrganizationId: %d - %s", c.IntId, c.Role.String())
}
