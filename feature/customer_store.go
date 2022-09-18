package feature

import (
	"context"
	"strings"
)

func (s Store) saveCustomers(ctx context.Context, cs ...customer) error {
	if len(cs) == 0 {
		// At least one customer must be given for the built query to be valid.
		return nil
	}

	var (
		qb   strings.Builder
		args []any
	)

	qb.WriteString(`INSERT INTO customer_features (id,feature_id,customer_id) VALUES`)
	for i, c := range cs {
		if 0 < i {
			qb.WriteRune(',')
		}
		qb.WriteString("(?,?,?)")
		args = append(args, c.ID, c.FeatureID, c.CustomerID)
	}

	_, err := s.db.ExecContext(
		ctx,
		qb.String(),
		args...,
	)
	return err
}
