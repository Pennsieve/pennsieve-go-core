package models

import (
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

func (c ServiceClaim) WithOrganizationClaim(claim organization.Claim) {
	c.Roles = append(c.Roles, ServiceRole{
		Type:   "organization_role",
		Id:     strconv.FormatInt(claim.IntId, 10),
		NodeId: claim.NodeId,
		Role:   claim.Role.AsOrganizationRole(),
	})
}

func (c ServiceClaim) WithDatasetClaim(claim dataset.Claim) {
	c.Roles = append(c.Roles, ServiceRole{
		Type:   "dataset_role",
		Id:     strconv.FormatInt(claim.IntId, 10),
		NodeId: claim.NodeId,
		Role:   strings.ToLower(claim.Role.String()),
	})
}
