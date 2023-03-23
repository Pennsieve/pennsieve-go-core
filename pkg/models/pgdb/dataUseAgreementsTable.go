package pgdb

import "time"

type DataUseAgreement struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
	IsDefault   bool      `json:"is_default"`
	Description string    `json:"description"`
}
