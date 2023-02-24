package authorizer

import (
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/core"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/permissions"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	"github.com/pennsieve/pennsieve-go-core/pkg/pgdb/models"
	log "github.com/sirupsen/logrus"
	"sort"
)

// Claims is an object containing claims and user info
type Claims struct {
	OrgClaim     organization.Claim
	DatasetClaim dataset.Claim
	UserClaim    user.Claim
}

// GetDatasetClaim returns the highest role that the user has for a given dataset.
// This method checks the roles of the dataset, the teams, and the specific user roles.
func GetDatasetClaim(db core.PostgresAPI, user *models.User, datasetNodeId string, organizationId int64) (*dataset.Claim, error) {

	// if user is super-admin
	if user.IsSuperAdmin {
		// USER IS A SUPER-ADMIN

		//TODO: HANDLE SPECIAL CASE
		log.Warn("Not handling super-user authorization at this point.")

	}

	// 1. Get Dataset Role and integer ID
	datasetQuery := fmt.Sprintf("SELECT id, role FROM \"%d\".datasets WHERE node_id='%s';", organizationId, datasetNodeId)

	var datasetId int64
	var maybeDatasetRole sql.NullString

	row := db.QueryRow(datasetQuery)
	err := row.Scan(
		&datasetId,
		&maybeDatasetRole)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			log.Error("No rows were returned!")
			return nil, err
		default:
			log.Error("Uknown Error while scanning dataset table: ", err)
			panic(err)
		}
	}

	// If maybeDatasetRole is set, include the role, otherwise use none-role
	datasetRole := dataset.None
	if maybeDatasetRole.Valid {
		var ok bool
		datasetRole, ok = dataset.RoleFromString(maybeDatasetRole.String)
		if !ok {
			log.Fatalln("Could not map Dataset Role from database string: ", maybeDatasetRole.String)
		}
	}

	// 2. Get Team Role
	teamPermission := fmt.Sprintf("\"%d\".dataset_team.role", organizationId)
	datasetTeam := fmt.Sprintf("\"%d\".dataset_team", organizationId)
	teamQueryStr := fmt.Sprintf("SELECT %s FROM pennsieve.team_user JOIN %s "+
		"ON pennsieve.team_user.team_id = %s.team_id "+
		"WHERE user_id=%d AND dataset_id=%d", teamPermission, datasetTeam, datasetTeam, user.Id, datasetId)

	// Get User Role
	userPermission := fmt.Sprintf("\"%d\".dataset_user.role", organizationId)
	datasetUser := fmt.Sprintf("\"%d\".dataset_user", organizationId)
	userQueryStr := fmt.Sprintf("SELECT %s FROM %s WHERE user_id=%d AND dataset_id=%d",
		userPermission, datasetUser, user.Id, datasetId)

	// Combine all queries in a single Union.
	fullQuery := teamQueryStr + " UNION " + userQueryStr + ";"

	rows, err := db.Query(fullQuery)
	if err != nil {
		return nil, err
	}

	roles := []dataset.Role{
		datasetRole,
	}
	for rows.Next() {
		var roleString string
		err = rows.Scan(
			&roleString)

		if err != nil {
			log.Error("ERROR: ", err)
		}

		role, ok := dataset.RoleFromString(roleString)
		if !ok {
			log.Fatalln("Could not map Dataset Role from database string.")
		}
		roles = append(roles, role)
	}

	// Sort roles by enum value --> first entry is the highest level of permission.
	sort.Slice(roles, func(i, j int) bool {
		return roles[i] > roles[j]
	})

	// return the maximum role that the user has.
	claim := dataset.Claim{
		Role:   roles[0],
		NodeId: datasetNodeId,
		IntId:  datasetId,
	}

	return &claim, nil

}

// ParseClaims creates a Claims object from a string map which is returned by the authorizer.
func ParseClaims(claims map[string]interface{}) *Claims {

	var orgClaim organization.Claim
	if val, ok := claims["org_claim"]; ok {
		orgClaims := val.(map[string]interface{})
		orgRole := int64(orgClaims["Role"].(float64))
		orgClaim = organization.Claim{
			Role:            models.DbPermission(orgRole),
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
