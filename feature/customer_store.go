package feature

import (
	"context"
	"strings"
	"time"
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

func (s Store) findCustomerFeaturesByTechnicalNames(ctx context.Context, customerID string, t time.Time, technicalNames ...string) ([]customerFeature, error) {
	if len(technicalNames) == 0 {
		return nil, nil
	}

	var (
		placeholders = make([]string, len(technicalNames))
		args         = make([]any, len(technicalNames)+2)
	)

	args[0], args[1] = customerID, t.Unix()
	for i, tn := range technicalNames {
		placeholders[i] = "?"
		args[i+2] = tn
	}

	query := `
	WITH only_customer_features (feature_id) AS (
		SELECT
			feature_id
		FROM customer_features
		WHERE customer_id = ?
	) SELECT
		f.technical_name,
		f.inverted,
		f.expires_on IS NOT NULL AND unixepoch(f.expires_on) < ? AS 'expired',
		ocf.feature_id IS NOT NULL AS 'customer_has_feature'
	FROM features f
		LEFT JOIN only_customer_features ocf ON f.id = ocf.feature_id
	WHERE f.technical_name IN (` + strings.Join(placeholders, ", ") + `)`

	rs, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	var cfs []customerFeature
	for rs.Next() {
		var cf customerFeature
		if err := rs.Scan(&cf.TechnicalName, &cf.Inverted, &cf.Expired, &cf.HasFeature); err != nil {
			return nil, err
		}
		cfs = append(cfs, cf)
	}

	if err := rs.Err(); err != nil {
		return nil, err
	}

	if err := rs.Close(); err != nil {
		return nil, err
	}

	return cfs, nil
}
