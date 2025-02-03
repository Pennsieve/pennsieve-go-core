package pgdb

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
)

// GetDefaultDataUseAgreement will return the default data use agreement for the organization.
// Returns (nil, sql.ErrNoRows) if no default data use agreement is found for the organization
func (q *Queries) GetDefaultDataUseAgreement(ctx context.Context, organizationId int) (*pgdb.DataUseAgreement, error) {
	query := fmt.Sprintf("SELECT id, name, body, created_at, is_default, description"+
		" FROM \"%d\".data_use_agreements where is_default = true;", organizationId)

	row := q.db.QueryRowContext(ctx, query)
	dataUseAgreement := pgdb.DataUseAgreement{}
	err := row.Scan(
		&dataUseAgreement.Id,
		&dataUseAgreement.Name,
		&dataUseAgreement.Body,
		&dataUseAgreement.CreatedAt,
		&dataUseAgreement.IsDefault,
		&dataUseAgreement.Description)

	if err != nil {
		return nil, err
	}

	return &dataUseAgreement, nil
}
