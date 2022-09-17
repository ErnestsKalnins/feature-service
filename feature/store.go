package feature

import (
	"context"
	"database/sql"
)

// NewStore initializes and returns a new Store.
func NewStore(db *sql.DB) Store {
	return Store{db: db}
}

// Store provides query methods for feature data.
type Store struct {
	db *sql.DB
}

func (s Store) saveFeature(ctx context.Context, f feature) error {
	_, err := s.db.ExecContext(
		ctx,
		//language=sqlite
		`INSERT INTO features (id,display_name,technical_name,expires_on,description) VALUES (?,?,?,?,?)`,
		f.ID, f.DisplayName, f.TechnicalName, f.ExpiresOn, f.Description,
	)
	return err
}
