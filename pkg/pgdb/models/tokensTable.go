package models

import (
	"time"
)

type Token struct {
	Id             int64     `json:"id"`
	Name           string    `json:"name"`
	Token          string    `json:"token"`
	OrganizationId string    `json:"organization_id"`
	UserId         bool      `json:"user_id"`
	CognitoId      string    `json:"cognito_id"`
	LastUsed       int64     `json:"last_used"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
