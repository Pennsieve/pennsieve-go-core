package authorizer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type ClaimResponse struct {
	Context map[string]interface{} `json:"context,omitempty"`
}

func TestClaims(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T){
		"Parse Claims":     testParseClaims,
		"Is Publisher":     testIsPublisher,
		"Is Not Publisher": testIsNotPublisher,
		"No Team Claims":   testNoTeamClaims,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func generated(withTeamClaims bool, withPublishingTeam bool) map[string]interface{} {
	orgClaim := make(map[string]interface{})
	orgClaim["Role"] = float64(16)
	orgClaim["IntId"] = float64(2001)
	orgClaim["NodeId"] = "N:organization:9e84e26c-1919-4864-9edc-b7082627601f"
	orgClaim["EnabledFeatures"] = nil

	datasetClaim := make(map[string]interface{})
	datasetClaim["Role"] = float64(32)
	datasetClaim["NodeId"] = "N:dataset:d83884a5-3034-4c08-86c0-9435757c5faa"
	datasetClaim["IntId"] = float64(2002)

	userClaim := make(map[string]interface{})
	userClaim["Id"] = float64(2003)
	userClaim["NodeId"] = "N:user:bb469ddb-82c5-405d-a700-7630bf49c388"
	userClaim["IsSuperAdmin"] = false

	var teamClaims []interface{}

	if withTeamClaims {
		teamClaim := make(map[string]interface{})
		teamClaim["IntId"] = float64(2004)
		teamClaim["Name"] = "Researchers"
		teamClaim["NodeId"] = "N:team:c20158f5-62c8-47b6-84c4-bc848b1a1313"
		teamClaim["Permission"] = float64(8)
		teamClaim["TeamType"] = ""

		teamClaims = append(teamClaims, teamClaim)
	}

	if withPublishingTeam {
		teamClaim := make(map[string]interface{})
		teamClaim["IntId"] = float64(2005)
		teamClaim["Name"] = "Publishers"
		teamClaim["NodeId"] = "N:team:0bb8fd6d-4560-413e-9488-d3fe8c8459c0"
		teamClaim["Permission"] = float64(8)
		teamClaim["TeamType"] = "publishers"

		teamClaims = append(teamClaims, teamClaim)
	}

	claims := map[string]interface{}{
		"user_claim":    userClaim,
		"org_claim":     orgClaim,
		"dataset_claim": datasetClaim,
		"team_claims":   teamClaims,
	}

	return claims
}

func testParseClaims(t *testing.T) {
	response := generated(true, false)
	claims := ParseClaims(response)
	assert.NotNil(t, claims.OrgClaim)
	assert.NotNil(t, claims.DatasetClaim)
	assert.NotNil(t, claims.UserClaim)
}

func testIsPublisher(t *testing.T) {
	response := generated(true, true)
	claims := ParseClaims(response)
	assert.NotNil(t, claims.OrgClaim)
	assert.NotNil(t, claims.DatasetClaim)
	assert.NotNil(t, claims.UserClaim)
	assert.True(t, IsPublisher(claims))
}

func testIsNotPublisher(t *testing.T) {
	response := generated(true, false)
	claims := ParseClaims(response)
	assert.NotNil(t, claims.OrgClaim)
	assert.NotNil(t, claims.DatasetClaim)
	assert.NotNil(t, claims.UserClaim)
	assert.False(t, IsPublisher(claims))
}

func testNoTeamClaims(t *testing.T) {
	response := generated(false, false)
	claims := ParseClaims(response)
	assert.NotNil(t, claims.OrgClaim)
	assert.NotNil(t, claims.DatasetClaim)
	assert.NotNil(t, claims.UserClaim)
	assert.False(t, IsPublisher(claims))
}
