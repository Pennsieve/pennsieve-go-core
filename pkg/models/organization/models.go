package organization

import (
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/pgdb/models"
)

// Claim combines the role of the user in the org, and the features in the organization.
type Claim struct {
	Role            models.DbPermission
	IntId           int64
	EnabledFeatures []models.FeatureFlags
}

func (c Claim) String() string {
	return fmt.Sprintf("OrganizationId: %d - %s", c.IntId, c.Role.String())
}
