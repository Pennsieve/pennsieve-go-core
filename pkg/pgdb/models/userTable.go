package models

type User struct {
	Id           int64  `json:"id"`
	NodeId       string `json:"node_id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	PreferredOrg int64  `json:"preferred_org_id"`
}
