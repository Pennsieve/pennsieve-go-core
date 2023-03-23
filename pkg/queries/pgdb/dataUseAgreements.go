package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
)

// GetDefaultDataUseAgreement will return the default data use agreement for the organization.
func (q *Queries) GetDefaultDataUseAgreement(ctx context.Context, organizationId int) (*pgdb.DataUseAgreement, error) {
	query := fmt.Sprintf("SELECT * FROM \"%d\".data_use_agreements where is_default = true;", organizationId)

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
		switch err {
		case sql.ErrNoRows:
			log.Error("No rows were returned!")
			return nil, err
		default:
			log.Error("Unknown Error while scanning data_use_agreements table: ", err)
			panic(err)
		}
	}

	return &dataUseAgreement, nil
}
