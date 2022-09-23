package feature

import (
	"context"
	"feature/pkg/slices"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

func (s Store) saveCustomers(ctx context.Context, cs ...customer) error {
	if len(cs) == 0 {
		// At least one customer must be given for the built query to be valid.
		return nil
	}

	query, args, err := goqu.Dialect("sqlite3").
		Insert(goqu.T("customer_features")).
		Rows(slices.Map(func(c customer) goqu.Record {
			return goqu.Record{
				"id":          c.ID,
				"feature_id":  c.FeatureID,
				"customer_id": c.CustomerID,
			}
		}, cs...)).
		Prepared(true).
		ToSQL()
	if err != nil {
		return fmt.Errorf("bad query: %w", err)
	}

	_, err = s.db.ExecContext(
		ctx,
		query,
		args...,
	)
	return err
}

func (s Store) findCustomerFeaturesByTechnicalNames(ctx context.Context, customerID string, t time.Time, technicalNames ...string) ([]customerFeature, error) {
	if len(technicalNames) == 0 {
		return nil, nil
	}

	query, args, err := goqu.Dialect("sqlite3").
		Select(
			goqu.I("f.technical_name"),
			goqu.I("f.inverted"),
			goqu.V(goqu.And(
				goqu.I("f.expires_on").IsNotNull(),
				goqu.I("f.expires_on").Lt(t),
			)).As("expired"),
			goqu.I("cf.feature_id").IsNotNull().As("customer_has_feature"),
		).
		From(goqu.T("features").As("f")).
		LeftJoin(
			goqu.T("customer_features").As("cf"),
			goqu.On(goqu.Ex{
				"f.id":           goqu.I("cf.feature_id"),
				"cf.customer_id": customerID,
			}),
		).
		Where(goqu.Ex{"f.technical_name": technicalNames}).
		Prepared(true).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("bad query: %w", err)
	}

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
