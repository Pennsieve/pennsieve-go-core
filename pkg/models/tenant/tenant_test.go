package tenant

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	orgId  = "3f2504e0-4f89-41d3-9a0c-0305e82c3301"
	userId = "9b2c1f7e-8d4a-4e6b-b1c3-2a5f6e7d8c90"
)

func TestOrgAndUser(t *testing.T) {
	org, err := Org(orgId)
	require.NoError(t, err)
	assert.Equal(t, OrganizationKind, org.Kind())
	assert.Equal(t, orgId, org.ID())
	assert.Equal(t, "ORG#"+orgId, org.String())
	assert.Equal(t, "N:organization:"+orgId, org.NodeId())

	user, err := User(userId)
	require.NoError(t, err)
	assert.Equal(t, UserKind, user.Kind())
	assert.Equal(t, "USER#"+userId, user.String())
	assert.Equal(t, "N:user:"+userId, user.NodeId())
}

func TestOnlyCanonicalUuids(t *testing.T) {
	for _, id := range []string{
		"",
		"not-a-uuid",
		"3F2504E0-4F89-41D3-9A0C-0305E82C3301",   // upper case
		"{3f2504e0-4f89-41d3-9a0c-0305e82c3301}", // braces
		"urn:uuid:3f2504e0-4f89-41d3-9a0c-0305e82c3301", // urn
		"3f2504e04f8941d39a0c0305e82c3301",              // no dashes
	} {
		_, err := Org(id)
		assert.Error(t, err, id)
		_, err = User(id)
		assert.Error(t, err, id)
	}
}

func TestFromNodeIds(t *testing.T) {
	org, err := FromOrgNodeId("N:organization:" + orgId)
	require.NoError(t, err)
	assert.Equal(t, "ORG#"+orgId, org.String())

	user, err := FromUserNodeId("N:user:" + userId)
	require.NoError(t, err)
	assert.Equal(t, "USER#"+userId, user.String())

	_, err = FromOrgNodeId("N:user:" + userId)
	assert.Error(t, err, "a user node id is not an organization")
	_, err = FromUserNodeId("N:organization:" + orgId)
	assert.Error(t, err, "an organization node id is not a user")
	_, err = FromOrgNodeId("INDEPENDENT")
	assert.Error(t, err)
}

func TestParse(t *testing.T) {
	org, err := Parse("ORG#" + orgId)
	require.NoError(t, err)
	assert.Equal(t, OrganizationKind, org.Kind())

	user, err := Parse("USER#" + userId)
	require.NoError(t, err)
	assert.Equal(t, UserKind, user.Kind())

	_, err = Parse("")
	assert.ErrorIs(t, err, ErrNoTenant)
	for _, s := range []string{"INDEPENDENT", "ORG#", "USER#", "TEAM#" + orgId, "org#" + orgId, orgId} {
		_, err := Parse(s)
		assert.Error(t, err, s)
	}
}

func TestZeroTenant(t *testing.T) {
	var zero Tenant
	assert.True(t, zero.IsZero())
	assert.Equal(t, "", zero.String())
	assert.Equal(t, "", zero.NodeId())

	_, err := json.Marshal(zero)
	assert.ErrorIs(t, err, ErrNoTenant)
	_, err = zero.MarshalDynamoDBAttributeValue()
	assert.ErrorIs(t, err, ErrNoTenant)
	assert.Panics(t, func() { Match(zero, func(string) int { return 0 }, func(string) int { return 0 }) })
}

func TestMatch(t *testing.T) {
	org, _ := Org(orgId)
	user, _ := User(userId)
	describe := func(t Tenant) string {
		return Match(t,
			func(orgNodeId string) string { return "org " + orgNodeId },
			func(userNodeId string) string { return "user " + userNodeId },
		)
	}
	assert.Equal(t, "org N:organization:"+orgId, describe(org))
	assert.Equal(t, "user N:user:"+userId, describe(user))
}

type record struct {
	Uuid   string `json:"uuid" dynamodbav:"uuid"`
	Tenant Tenant `json:"tenant" dynamodbav:"tenant"`
}

func TestJSONRoundTrip(t *testing.T) {
	org, _ := Org(orgId)
	data, err := json.Marshal(record{Uuid: "n1", Tenant: org})
	require.NoError(t, err)
	assert.JSONEq(t, `{"uuid":"n1","tenant":"ORG#`+orgId+`"}`, string(data))

	var back record
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, org, back.Tenant)

	assert.Error(t, json.Unmarshal([]byte(`{"tenant":"INDEPENDENT"}`), &back))
	assert.Error(t, json.Unmarshal([]byte(`{"tenant":7}`), &back))
}

func TestDynamoDBRoundTrip(t *testing.T) {
	user, _ := User(userId)
	item, err := attributevalue.MarshalMap(record{Uuid: "n1", Tenant: user})
	require.NoError(t, err)
	assert.Equal(t, &types.AttributeValueMemberS{Value: "USER#" + userId}, item["tenant"])

	var back record
	require.NoError(t, attributevalue.UnmarshalMap(item, &back))
	assert.Equal(t, user, back.Tenant)

	_, err = attributevalue.MarshalMap(record{Uuid: "n1"})
	assert.ErrorIs(t, err, ErrNoTenant, "a record without a tenant cannot be saved")

	var legacy record
	require.NoError(t, attributevalue.UnmarshalMap(map[string]types.AttributeValue{
		"uuid": &types.AttributeValueMemberS{Value: "n1"},
	}, &legacy))
	assert.True(t, legacy.Tenant.IsZero(), "a record from before tenants reads as the zero Tenant")

	err = attributevalue.UnmarshalMap(map[string]types.AttributeValue{
		"tenant": &types.AttributeValueMemberS{Value: "INDEPENDENT"},
	}, &back)
	assert.Error(t, err)
	err = attributevalue.UnmarshalMap(map[string]types.AttributeValue{
		"tenant": &types.AttributeValueMemberN{Value: "1"},
	}, &back)
	assert.Error(t, err)
}
