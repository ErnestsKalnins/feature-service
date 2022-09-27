package feature

import (
	"context"
	"database/sql"
	"errors"
	"feature/pkg/sqlx"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// NewStore initializes and returns a new Store.
func NewStore(db *sql.DB) Store {
	return Store{db: db}
}

// Store provides query methods for feature data.
type Store struct {
	db db
}

// db contains methods common to *sql.DB and *sql.Tx.
type db interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

func (s Store) beginTx(ctx context.Context, opts *sql.TxOptions) (*Store, func() error, func() error, error) {
	switch v := s.db.(type) {
	case *sql.DB:
		tx, err := v.BeginTx(ctx, opts)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("begin transaction: %w", err)
		}
		return &Store{db: tx}, tx.Commit, tx.Rollback, nil
	case *sql.Tx:
		// Transaction already in progress, return self. We return a noop Commit and
		// Rollback func in order to allow the first call to `beginTx` to control the
		// transaction.
		noop := func() error { return nil }
		return &s, noop, noop, nil
	default:
		return nil, nil, nil, fmt.Errorf("unexpected db type: %T", v)
	}
}

func (s Store) findAllFeatures(ctx context.Context) ([]feature, error) {
	rs, err := s.db.QueryContext(
		ctx,
		//language=sqlite
		`SELECT id,display_name,technical_name,expires_on,description,inverted,created_at,updated_at FROM features`,
	)
	if err != nil {
		return nil, err
	}

	var fs []feature
	for rs.Next() {
		var fr featureRow
		if err := rs.Scan(
			&fr.ID,
			&fr.DisplayName,
			&fr.TechnicalName,
			&fr.ExpiresOn,
			&fr.Description,
			&fr.Inverted,
			&fr.CreatedAt,
			&fr.UpdatedAt,
		); err != nil {
			return nil, err
		}
		fs = append(fs, fr.toFeature())
	}

	if err := rs.Err(); err != nil {
		return nil, err
	}

	if err := rs.Close(); err != nil {
		return nil, err
	}

	return fs, nil
}

func (s Store) findFeature(ctx context.Context, id uuid.UUID) (*feature, error) {
	r := s.db.QueryRowContext(
		ctx,
		//language=sqlite
		`SELECT display_name,technical_name,expires_on,description,inverted,created_at,updated_at FROM features WHERE id=?`,
		id,
	)

	fr := featureRow{ID: id}
	if err := r.Scan(
		&fr.DisplayName,
		&fr.TechnicalName,
		&fr.ExpiresOn,
		&fr.Description,
		&fr.Inverted,
		&fr.CreatedAt,
		&fr.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errFeatureNotFound{id: id}
		}
		return nil, err
	}

	f := fr.toFeature()
	return &f, nil
}

func (s Store) findFeatureWithClients(ctx context.Context, id uuid.UUID) (*feature, error) {
	r := s.db.QueryRowContext(
		ctx,
		//language=sqlite
		`
		SELECT
			f.display_name,
			f.technical_name,
			f.expires_on,
			f.description,
			f.inverted,
			f.created_at,
			f.updated_at,
			CASE WHEN cf.customer_id IS NOT NULL THEN json_group_array(cf.customer_id)
		END AS 'customer_ids'
		FROM features f
		LEFT JOIN customer_features cf ON f.id = cf.feature_id
		WHERE f.id=?`,
		id,
	)

	fr := featureRow{ID: id}
	if err := r.Scan(
		&fr.DisplayName,
		&fr.TechnicalName,
		&fr.ExpiresOn,
		&fr.Description,
		&fr.Inverted,
		&fr.CreatedAt,
		&fr.UpdatedAt,
		&fr.CustomerIDs,
	); err != nil {
		return nil, err
	}

	if err := r.Err(); err != nil {
		return nil, err
	}

	f := fr.toFeature()
	return &f, nil
}

func (s Store) saveFeature(ctx context.Context, f feature) error {
	r := featureToRow(f)
	_, err := s.db.ExecContext(
		ctx,
		//language=sqlite
		`INSERT INTO features (id,display_name,technical_name,expires_on,description,inverted,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)`,
		r.ID, r.DisplayName, r.TechnicalName, r.ExpiresOn, r.Description, r.Inverted, r.CreatedAt, r.UpdatedAt,
	)
	return err
}

func (s Store) updateFeature(ctx context.Context, lastUpdatedAt time.Time, f feature) error {
	r := featureToRow(f)
	res, err := s.db.ExecContext(
		ctx,
		//language=sqlite
		`UPDATE features SET display_name=?, technical_name=?, expires_on=?, description=?, inverted=?, updated_at=? WHERE id=? AND unixepoch(updated_at)=?`,
		r.DisplayName, r.TechnicalName, r.ExpiresOn, r.Description, r.Inverted, r.UpdatedAt, r.ID, lastUpdatedAt.Unix(),
	)
	if err != nil {
		return err
	}

	rs, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rs == 0 {
		// Unfortunately, there is no way to determine whether update failed due to the
		// row not existing, or due to updating based on stale data.
		return errFeatureNotFound{id: f.ID}
	}
	return nil
}

type errFeatureNotFound struct {
	id uuid.UUID
}

func (e errFeatureNotFound) Error() string {
	return fmt.Sprintf("feature %s does not exist", e.id)
}

func (e errFeatureNotFound) Code() int {
	return http.StatusNotFound
}

func (s Store) deleteFeature(ctx context.Context, featureID uuid.UUID) error {
	_, err := s.db.ExecContext(
		ctx,
		//language=sqlite
		`DELETE FROM features WHERE id=?`,
		featureID,
	)
	return err
}

func featureToRow(f feature) featureRow {
	r := featureRow{
		ID:            f.ID,
		TechnicalName: f.TechnicalName,
		Inverted:      f.Inverted,
		CreatedAt:     f.CreatedAt.UTC(),
		UpdatedAt:     f.UpdatedAt.UTC(),
	}
	if f.DisplayName != nil {
		r.DisplayName = sql.NullString{String: *f.DisplayName, Valid: true}
	}
	if f.ExpiresOn != nil {
		r.ExpiresOn = sql.NullTime{Time: f.ExpiresOn.UTC(), Valid: true}
	}
	if f.Description != nil {
		r.Description = sql.NullString{String: *f.Description, Valid: true}
	}
	return r
}

type featureRow struct {
	ID            uuid.UUID
	DisplayName   sql.NullString
	TechnicalName string
	ExpiresOn     sql.NullTime
	Description   sql.NullString
	Inverted      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CustomerIDs   sqlx.JSONArray[string]
}

func (r featureRow) toFeature() feature {
	f := feature{
		ID:            r.ID,
		TechnicalName: r.TechnicalName,
		Inverted:      r.Inverted,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
	if r.DisplayName.Valid {
		f.DisplayName = &r.DisplayName.String
	}
	if r.ExpiresOn.Valid {
		f.ExpiresOn = &r.ExpiresOn.Time
	}
	if r.Description.Valid {
		f.Description = &r.Description.String
	}
	if 0 < len(r.CustomerIDs) {
		f.CustomerIDs = r.CustomerIDs
	}
	return f
}
