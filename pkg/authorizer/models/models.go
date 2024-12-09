package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"strconv"
	"strings"
)

type ServiceRole struct {
	Type   string `json:"type"`
	Id     string `json:"id"`
	NodeId string `json:"node_id"`
	Role   string `json:"role"`
}

type ServiceClaim struct {
	Type      string        `json:"type"`
	IssuedAt  string        `json:"iat"`
	ExpiresAt string        `json:"exp"`
	Roles     []ServiceRole `json:"roles"`
}

type ServiceToken struct {
	Value string `json:"value"`
}

func (c ServiceClaim) WithOrganizationClaim(claim *organization.Claim) ServiceClaim {
	c.Roles = append(c.Roles, ServiceRole{
		Type:   "organization_role",
		Id:     strconv.FormatInt(claim.IntId, 10),
		NodeId: claim.NodeId,
		Role:   claim.Role.AsRoleString(),
	})
	return c
}

func (c ServiceClaim) WithDatasetClaim(claim *dataset.Claim) ServiceClaim {
	c.Roles = append(c.Roles, ServiceRole{
		Type:   "dataset_role",
		Id:     strconv.FormatInt(claim.IntId, 10),
		NodeId: claim.NodeId,
		Role:   strings.ToLower(claim.Role.String()),
	})
	return c
}

func (c ServiceClaim) AsToken(key string) (*ServiceToken, error) {
	var (
		err          error
		secret       []byte
		token        *jwt.Token
		signedString string
	)
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":   c.IssuedAt,
		"exp":   c.ExpiresAt,
		"type":  c.Type,
		"roles": c.Roles,
	})
	secret = []byte(key)
	signedString, err = token.SignedString(secret)
	if err != nil {
		return nil, err
	}
	return &ServiceToken{Value: signedString}, nil
}
