package authorizer

import (
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/authorizer/models"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/permissions"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const LabelUserClaim = "user_claim"
const LabelOrganizationClaim = "org_claim"
const LabelTeamClaims = "team_claims"
const LabelDatasetClaim = "dataset_claim"
const LabelServiceClaim = "service_claim"

// Claims is an object containing claims and user info
type Claims struct {
	OrgClaim     *organization.Claim
	DatasetClaim *dataset.Claim
	UserClaim    *user.Claim
	TeamClaims   []teamUser.Claim
}

func (c *Claims) String() string {
	return fmt.Sprintf("{ OrgClaim: %s, DatasetClaim: %s, UserClaim: %s, TeamClaims: %s }",
		c.OrgClaim, c.DatasetClaim, c.UserClaim, c.TeamClaims)
}

// ParseClaims creates a Claims object from a string map which is returned by the authorizer.
func ParseClaims(claims map[string]interface{}) *Claims {
	log.WithFields(log.Fields{"service": "Authorizer", "function": "ParseClaims()", "claims": claims}).Debug()

	var orgClaim *organization.Claim
	if val, ok := claims[LabelOrganizationClaim]; ok {
		orgClaims := val.(map[string]interface{})
		orgRole := int64(orgClaims["Role"].(float64))
		orgClaim = &organization.Claim{
			Role:            pgdb.DbPermission(orgRole),
			IntId:           int64(orgClaims["IntId"].(float64)),
			NodeId:          orgClaims["NodeId"].(string),
			EnabledFeatures: nil,
		}
	}

	var datasetClaim *dataset.Claim
	if val, ok := claims[LabelDatasetClaim]; ok {
		if val != nil {
			datasetClaims := val.(map[string]interface{})
			datasetRole := int64(datasetClaims["Role"].(float64))
			datasetClaim = &dataset.Claim{
				Role:   role.Role(datasetRole),
				NodeId: datasetClaims["NodeId"].(string),
				IntId:  int64(datasetClaims["IntId"].(float64)),
			}
		}
	}

	var userClaim *user.Claim
	if val, ok := claims[LabelUserClaim]; ok {
		if val != nil {
			userClaims := val.(map[string]interface{})
			userClaim = &user.Claim{
				Id:           int64(userClaims["Id"].(float64)),
				NodeId:       userClaims["NodeId"].(string),
				IsSuperAdmin: userClaims["IsSuperAdmin"].(bool),
			}
		}
	}

	var teamClaims []teamUser.Claim
	if val, ok := claims[LabelTeamClaims]; ok {
		if val != nil {
			tcs := val.([]interface{})
			for _, item := range tcs {
				tc := item.(map[string]interface{})
				teamClaim := teamUser.Claim{
					IntId:      int64(tc["IntId"].(float64)),
					Name:       tc["Name"].(string),
					NodeId:     tc["NodeId"].(string),
					Permission: pgdb.DbPermission(int64(tc["Permission"].(float64))),
					TeamType:   tc["TeamType"].(string),
				}
				teamClaims = append(teamClaims, teamClaim)
			}
		}
	}

	parsedClaims := Claims{
		OrgClaim:     orgClaim,
		DatasetClaim: datasetClaim,
		UserClaim:    userClaim,
		TeamClaims:   teamClaims,
	}
	log.WithFields(log.Fields{"service": "Authorizer", "function": "ParseClaims()", "parsedClaims": parsedClaims}).Debug()

	return &parsedClaims
}

// HasOrgRole returns true if this claim contains an OrgClaim with permissions sufficient to satisfy the given requiredOrgRole
func (c *Claims) HasOrgRole(requiredOrgRole role.Role) bool {
	if c == nil {
		return false
	}
	return c.OrgClaim.HasRole(requiredOrgRole)
}

// HasRole returns a boolean indicating whether the given Claims contain a dataset.Claim with permissions sufficient to
// satisfy the given permissions.DatasetPermission
func HasRole(claims Claims, permission permissions.DatasetPermission) bool {

	//hasOrgRole := claims.orgClaim.Role >= dbTable.Delete

	hasValidPermissions := permissions.HasDatasetPermission(claims.DatasetClaim.Role, permission)

	return hasValidPermissions

}

// HasOrgRole returns true if the given *Claims contains an OrgClaim with permissions sufficient to satisfy the given requiredOrgRole
func HasOrgRole(claims *Claims, requiredOrgRole role.Role) bool {
	return claims.HasOrgRole(requiredOrgRole)
}

// IsPublisher returns a boolean indicating whether the user is on the Publishing team
func IsPublisher(claims *Claims) bool {
	isPublisher := false

	for _, claim := range claims.TeamClaims {
		if claim.TeamType == "publishers" {
			isPublisher = true
			break
		}
	}

	return isPublisher
}

func GenerateServiceClaim(duration time.Duration) models.ServiceClaim {
	issuedTime := time.Now().Unix()
	expiresAt := issuedTime + duration.Milliseconds()/1000
	return models.ServiceClaim{
		Type:      LabelServiceClaim,
		IssuedAt:  strconv.FormatInt(issuedTime, 10),
		ExpiresAt: strconv.FormatInt(expiresAt, 10),
	}
}
