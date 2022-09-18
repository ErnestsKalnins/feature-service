package feature

import (
	"context"
	"database/sql"
	"errors"
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

func (s Store) findFeature(ctx context.Context, id uuid.UUID) (*feature, error) {
	r := s.db.QueryRowContext(
		ctx,
		//language=sqlite
		`SELECT display_name,technical_name,expires_on,description,inverted,created_at,updated_at FROM features WHERE id=?`,
		id,
	)

	f := feature{ID: id}
	if err := r.Scan(
		&f.DisplayName,
		&f.TechnicalName,
		&f.ExpiresOn,
		&f.Description,
		&f.Inverted,
		&f.CreatedAt,
		&f.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errFeatureNotFound{id: id}
		}
		return nil, err
	}
	return &f, nil
}

func (s Store) saveFeature(ctx context.Context, f feature) error {
	_, err := s.db.ExecContext(
		ctx,
		//language=sqlite
		`INSERT INTO features (id,display_name,technical_name,expires_on,description,inverted,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)`,
		f.ID, f.DisplayName, f.TechnicalName, f.ExpiresOn, f.Description, f.Inverted, f.CreatedAt, f.UpdatedAt,
	)
	return err
}

func (s Store) updateFeature(ctx context.Context, lastUpdatedAt time.Time, f feature) error {
	res, err := s.db.ExecContext(
		ctx,
		//language=sqlite
		`UPDATE features SET display_name=?, technical_name=?, expires_on=?, description=?, inverted=?, updated_at=? WHERE id=? AND updated_at=?`,
		f.DisplayName, f.TechnicalName, f.ExpiresOn, f.Description, f.Inverted, f.UpdatedAt, f.ID, lastUpdatedAt,
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
