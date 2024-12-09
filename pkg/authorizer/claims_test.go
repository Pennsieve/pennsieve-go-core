package authorizer

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

type ClaimResponse struct {
	Context map[string]interface{} `json:"context,omitempty"`
}

func TestClaims(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T){
		"Parse Claims":                        testParseClaims,
		"Is Publisher":                        testIsPublisher,
		"Is Not Publisher":                    testIsNotPublisher,
		"No Team Claims":                      testNoTeamClaims,
		"Generate a Service Claim":            testGenerateServiceClaim,
		"Generate a Service Claim with roles": testGenerateServiceClaimWithRoles,
		"Service Claim as Token":              testServiceClaimAsToken,
		"Nil Claims have no org role":         testNilClaimsHaveNoOrgRole,
		"HasOrgRole":                          testHasOrgRole,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func TestClaimBugs(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T){
		"Fail to parse Teams Claim":      testBugFailToParseTeamsClaim,
		"Successfully parse Team Claims": testBugSuccessfullyParseTeamClaims,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func generated(withTeamClaims bool, withPublishingTeam bool) map[string]interface{} {
	orgClaimV := make(map[string]interface{})
	orgClaimV["Role"] = float64(16)
	orgClaimV["IntId"] = float64(2001)
	orgClaimV["NodeId"] = "N:organization:9e84e26c-1919-4864-9edc-b7082627601f"
	orgClaimV["EnabledFeatures"] = nil

	datasetClaimV := make(map[string]interface{})
	datasetClaimV["Role"] = float64(32)
	datasetClaimV["NodeId"] = "N:dataset:d83884a5-3034-4c08-86c0-9435757c5faa"
	datasetClaimV["IntId"] = float64(2002)

	userClaimV := make(map[string]interface{})
	userClaimV["Id"] = float64(2003)
	userClaimV["NodeId"] = "N:user:bb469ddb-82c5-405d-a700-7630bf49c388"
	userClaimV["IsSuperAdmin"] = false

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
		LabelUserClaim:         userClaimV,
		LabelOrganizationClaim: orgClaimV,
		LabelDatasetClaim:      datasetClaimV,
		LabelTeamClaims:        teamClaims,
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

var duration time.Duration = 5 * time.Minute

func orgClaim() *organization.Claim {
	return &organization.Claim{
		Role:            pgdb.Owner,
		IntId:           367,
		NodeId:          "N:organization:06c8002d-477a-45e9-ae0d-06f4b218628f",
		EnabledFeatures: nil,
	}
}

func datasetClaim() *dataset.Claim {
	return &dataset.Claim{
		Role:   role.Owner,
		NodeId: "N:dataset:ca645a17-fb55-4afd-aff8-7e0078b4523f",
		IntId:  86,
	}
}

func testGenerateServiceClaim(t *testing.T) {
	claim := GenerateServiceClaim(duration)
	assert.NotNil(t, claim)
	// verify that IssuedAt and ExpiresAt are non-zero
	issuedAt, _ := strconv.Atoi(claim.IssuedAt)
	assert.Greater(t, issuedAt, 0)
	expiresAt, _ := strconv.Atoi(claim.ExpiresAt)
	assert.Greater(t, expiresAt, 0)
	// verify that ExpiresAt is "later than" IssuedAt
	assert.Greater(t, expiresAt, issuedAt)
	// verify that the validity period is what was asked for
	period, err := time.ParseDuration(fmt.Sprintf("%ds", expiresAt-issuedAt))
	assert.NoError(t, err)
	assert.Equal(t, duration, period)
}

func testGenerateServiceClaimWithRoles(t *testing.T) {
	claim := GenerateServiceClaim(duration).WithOrganizationClaim(orgClaim()).WithDatasetClaim(datasetClaim())
	assert.NotNil(t, claim)
	// verify that the provided claims were included
	assert.Equal(t, 2, len(claim.Roles))
}

func testServiceClaimAsToken(t *testing.T) {
	token, err := GenerateServiceClaim(5 * time.Minute).WithOrganizationClaim(orgClaim()).WithDatasetClaim(datasetClaim()).AsToken("secret")
	assert.NoError(t, err)
	assert.NotNil(t, token)
	// verify that the encoded token value is present
	assert.Greater(t, len(token.Value), 0)
	// verify that the encoded JWT has the requisite 3 parts
	jwtParts := strings.Split(token.Value, ".")
	assert.Len(t, jwtParts, 3)
}

var allRoles = []role.Role{role.None, role.Guest, role.Viewer, role.Editor, role.Manager, role.Owner}

func testNilClaimsHaveNoOrgRole(t *testing.T) {
	var claims *Claims = nil
	for _, requireRole := range allRoles {
		assert.False(t, claims.HasOrgRole(requireRole))
		assert.False(t, HasOrgRole(claims, requireRole))
	}
}

func userClaim() *user.Claim {
	return &user.Claim{
		Id:           rand.Int63n(50),
		NodeId:       fmt.Sprintf("N:user:%s", uuid.NewString()),
		IsSuperAdmin: false,
	}
}

func testHasOrgRole(t *testing.T) {
	for _, testParams := range []struct {
		name           string
		role           pgdb.DbPermission
		expectedToHave map[role.Role]bool
	}{
		{
			"NoPermission",
			pgdb.NoPermission,
			map[role.Role]bool{role.None: true},
		},
		{
			"Guest",
			pgdb.Guest,
			map[role.Role]bool{role.None: true, role.Guest: true},
		},
		{
			"Read",
			pgdb.Read,
			map[role.Role]bool{role.None: true, role.Guest: true, role.Viewer: true},
		},
		{
			"Write",
			pgdb.Write,
			map[role.Role]bool{role.None: true, role.Guest: true, role.Viewer: true, role.Editor: true},
		},
		{
			"Delete",
			pgdb.Delete,
			map[role.Role]bool{role.None: true, role.Guest: true, role.Viewer: true, role.Editor: true},
		},
		{
			"Administer",
			pgdb.Administer,
			map[role.Role]bool{role.None: true, role.Guest: true, role.Viewer: true, role.Editor: true, role.Manager: true},
		},
		{
			"Owner",
			pgdb.Owner,
			map[role.Role]bool{role.None: true, role.Guest: true, role.Viewer: true, role.Editor: true, role.Manager: true, role.Owner: true},
		},
	} {
		t.Run(testParams.name, func(t *testing.T) {
			actualOrgClaim := &organization.Claim{
				Role:            testParams.role,
				IntId:           rand.Int63n(50),
				NodeId:          fmt.Sprintf("N:organization:%s", uuid.NewString()),
				EnabledFeatures: nil,
			}
			claims := &Claims{
				OrgClaim:     actualOrgClaim,
				DatasetClaim: datasetClaim(),
				UserClaim:    userClaim(),
				TeamClaims:   nil,
			}
			for _, requiredRole := range allRoles {
				expected := testParams.expectedToHave[requiredRole]
				actual := claims.HasOrgRole(requiredRole)
				assert.Equal(t, expected, actual)

				actual = HasOrgRole(claims, requiredRole)
				assert.Equal(t, expected, actual)
			}

		})
	}
}

func testBugFailToParseTeamsClaim(t *testing.T) {
	badClaimSource := "{\n        \"org_claim\": {\n            \"EnabledFeatures\": [\n                {\n                    \"created_at\": \"2023-08-23T22:50:03.381715Z\",\n                    \"enabled\": true,\n                    \"feature\": \"publishing50_feature\",\n                    \"organization_id\": 39,\n                    \"updated_at\": \"2023-08-23T22:50:03.381715Z\"\n                }\n            ],\n            \"IntId\": 39,\n            \"NodeId\": \"N:organization:7c2de0a6-5972-4138-99ad-cc0aff0fb67f\",\n            \"Role\": 32\n        },\n        \"teams_claim\": [\n            {\n                \"IntId\": 91,\n                \"Name\": \"Publishers\",\n                \"NodeId\": \"N:team:3a616648-9ad0-4809-be63-8615f08babad\",\n                \"Permission\": 16,\n                \"TeamType\": \"publishers\"\n            }\n        ],\n        \"user_claim\": {\n            \"Id\": 177,\n            \"IsSuperAdmin\": true,\n            \"NodeId\": \"N:user:61e7c1cf-a836-421b-b919-a2309402c9d6\"\n        }\n    }"
	input := []byte(badClaimSource)
	var claims map[string]interface{}
	err := json.Unmarshal(input, &claims)
	assert.NoError(t, err)
	parsedClaims := ParseClaims(claims)
	assert.True(t, len(parsedClaims.TeamClaims) == 0)
}

func testBugSuccessfullyParseTeamClaims(t *testing.T) {
	goodClaimSource := "{\n        \"org_claim\": {\n            \"EnabledFeatures\": [\n                {\n                    \"created_at\": \"2023-08-23T22:50:03.381715Z\",\n                    \"enabled\": true,\n                    \"feature\": \"publishing50_feature\",\n                    \"organization_id\": 39,\n                    \"updated_at\": \"2023-08-23T22:50:03.381715Z\"\n                }\n            ],\n            \"IntId\": 39,\n            \"NodeId\": \"N:organization:7c2de0a6-5972-4138-99ad-cc0aff0fb67f\",\n            \"Role\": 32\n        },\n        \"team_claims\": [\n            {\n                \"IntId\": 91,\n                \"Name\": \"Publishers\",\n                \"NodeId\": \"N:team:3a616648-9ad0-4809-be63-8615f08babad\",\n                \"Permission\": 16,\n                \"TeamType\": \"publishers\"\n            }\n        ],\n        \"user_claim\": {\n            \"Id\": 177,\n            \"IsSuperAdmin\": true,\n            \"NodeId\": \"N:user:61e7c1cf-a836-421b-b919-a2309402c9d6\"\n        }\n    }"
	input := []byte(goodClaimSource)
	var claims map[string]interface{}
	err := json.Unmarshal(input, &claims)
	assert.NoError(t, err)
	parsedClaims := ParseClaims(claims)
	assert.True(t, len(parsedClaims.TeamClaims) == 1)
}
