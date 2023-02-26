package authorizer

import (
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/permissions"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
)

// Claims is an object containing claims and user info
type Claims struct {
	OrgClaim     organization.Claim
	DatasetClaim dataset.Claim
	UserClaim    user.Claim
}

// ParseClaims creates a Claims object from a string map which is returned by the authorizer.
func ParseClaims(claims map[string]interface{}) *Claims {

	var orgClaim organization.Claim
	if val, ok := claims["org_claim"]; ok {
		orgClaims := val.(map[string]interface{})
		orgRole := int64(orgClaims["Role"].(float64))
		orgClaim = organization.Claim{
			Role:            pgdb.DbPermission(orgRole),
			IntId:           int64(orgClaims["IntId"].(float64)),
			EnabledFeatures: nil,
		}
	}

	var datasetClaim dataset.Claim
	if val, ok := claims["dataset_claim"]; ok {
		if val != nil {
			datasetClaims := val.(map[string]interface{})
			datasetRole := int64(datasetClaims["Role"].(float64))
			datasetClaim = dataset.Claim{
				Role:   dataset.Role(datasetRole),
				NodeId: datasetClaims["NodeId"].(string),
				IntId:  int64(datasetClaims["IntId"].(float64)),
			}
		}
	}

	var userClaim user.Claim
	if val, ok := claims["user_claim"]; ok {
		if val != nil {
			userClaims := val.(map[string]interface{})
			userClaim = user.Claim{
				Id:           int64(userClaims["Id"].(float64)),
				NodeId:       userClaims["NodeId"].(string),
				IsSuperAdmin: userClaims["IsSuperAdmin"].(bool),
			}
		}
	}

	returnedClaims := Claims{
		OrgClaim:     orgClaim,
		DatasetClaim: datasetClaim,
		UserClaim:    userClaim,
	}

	return &returnedClaims

}

// HasRole returns a boolean indicating whether the user has the correct permissions.
func HasRole(claims Claims, permission permissions.DatasetPermission) bool {

	//hasOrgRole := claims.orgClaim.Role >= dbTable.Delete

	hasValidPermissions := permissions.HasDatasetPermission(claims.DatasetClaim.Role, permission)

	return hasValidPermissions

}
