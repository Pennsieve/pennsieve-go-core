// Package tenant names who owns, and pays for, a resource: an organization
// or a user, never neither.
//
// A tenant is stored as one string, ORG#{uuid} or USER#{uuid}. It is never
// empty, so it can be a DynamoDB key without a sparse-index question, and it
// is prefix-queryable. The zero Tenant is not a tenant: marshalling it fails,
// so a record cannot be saved without one.
//
// A tenant is not an access grant. It answers who owns and pays for
// something; whether a caller may use it is a separate check.
package tenant

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type Kind int

const (
	kindNone Kind = iota
	OrganizationKind
	UserKind
)

func (k Kind) String() string {
	switch k {
	case OrganizationKind:
		return "organization"
	case UserKind:
		return "user"
	default:
		return "none"
	}
}

const (
	orgPrefix  = "ORG#"
	userPrefix = "USER#"

	orgNodeIdPrefix  = "N:organization:"
	userNodeIdPrefix = "N:user:"
)

// ErrNoTenant is returned when the zero Tenant is marshalled, or an empty
// value is parsed.
var ErrNoTenant = errors.New("tenant: no tenant")

type Tenant struct {
	kind Kind
	id   string // canonical lowercase uuid
}

var (
	_ attributevalue.Marshaler   = Tenant{}
	_ attributevalue.Unmarshaler = (*Tenant)(nil)
	_ json.Marshaler             = Tenant{}
	_ json.Unmarshaler           = (*Tenant)(nil)
)

// Org is the tenant for the organization with this uuid.
func Org(id string) (Tenant, error) { return newTenant(OrganizationKind, id) }

// User is the tenant for the user with this uuid.
func User(id string) (Tenant, error) { return newTenant(UserKind, id) }

// FromOrgNodeId is the tenant for an organization node id
// (N:organization:{uuid}).
func FromOrgNodeId(nodeId string) (Tenant, error) {
	id, ok := strings.CutPrefix(nodeId, orgNodeIdPrefix)
	if !ok {
		return Tenant{}, fmt.Errorf("tenant: %q is not an organization node id", nodeId)
	}
	return Org(id)
}

// FromUserNodeId is the tenant for a user node id (N:user:{uuid}).
func FromUserNodeId(nodeId string) (Tenant, error) {
	id, ok := strings.CutPrefix(nodeId, userNodeIdPrefix)
	if !ok {
		return Tenant{}, fmt.Errorf("tenant: %q is not a user node id", nodeId)
	}
	return User(id)
}

// Parse reads the stored form, ORG#{uuid} or USER#{uuid}.
func Parse(s string) (Tenant, error) {
	if s == "" {
		return Tenant{}, ErrNoTenant
	}
	if id, ok := strings.CutPrefix(s, orgPrefix); ok {
		return Org(id)
	}
	if id, ok := strings.CutPrefix(s, userPrefix); ok {
		return User(id)
	}
	return Tenant{}, fmt.Errorf("tenant: %q is neither ORG#{uuid} nor USER#{uuid}", s)
}

func newTenant(kind Kind, id string) (Tenant, error) {
	// Only the canonical form, so one tenant has one stored value.
	u, err := uuid.Parse(id)
	if err != nil || u.String() != id {
		return Tenant{}, fmt.Errorf("tenant: %q is not a canonical lowercase uuid", id)
	}
	return Tenant{kind: kind, id: id}, nil
}

// IsZero reports whether t is the zero Tenant, for example a record read
// before it had a tenant.
func (t Tenant) IsZero() bool { return t.kind == kindNone }

func (t Tenant) Kind() Kind { return t.kind }

// ID is the organization's or user's uuid.
func (t Tenant) ID() string { return t.id }

// NodeId is the organization's or user's node id, or "" for the zero Tenant.
func (t Tenant) NodeId() string {
	switch t.kind {
	case OrganizationKind:
		return orgNodeIdPrefix + t.id
	case UserKind:
		return userNodeIdPrefix + t.id
	default:
		return ""
	}
}

// String is the stored form, or "" for the zero Tenant.
func (t Tenant) String() string {
	switch t.kind {
	case OrganizationKind:
		return orgPrefix + t.id
	case UserKind:
		return userPrefix + t.id
	default:
		return ""
	}
}

// Match calls org or user with the tenant's node id, so a caller has to
// handle both. It panics on the zero Tenant; check IsZero first where a
// record may predate tenants.
func Match[T any](t Tenant, org func(orgNodeId string) T, user func(userNodeId string) T) T {
	switch t.kind {
	case OrganizationKind:
		return org(t.NodeId())
	case UserKind:
		return user(t.NodeId())
	default:
		panic(ErrNoTenant)
	}
}

func (t Tenant) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return nil, ErrNoTenant
	}
	return json.Marshal(t.String())
}

func (t *Tenant) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("tenant: %w", err)
	}
	parsed, err := Parse(s)
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

func (t Tenant) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	if t.IsZero() {
		return nil, ErrNoTenant
	}
	return &types.AttributeValueMemberS{Value: t.String()}, nil
}

func (t *Tenant) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	s, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return fmt.Errorf("tenant: expected a string attribute, got %T", av)
	}
	parsed, err := Parse(s.Value)
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}
