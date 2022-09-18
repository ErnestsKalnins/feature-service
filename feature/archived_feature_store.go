package feature

import "context"

func (s Store) saveArchivedFeature(ctx context.Context, f feature) error {
	_, err := s.db.ExecContext(
		ctx,
		//language=sqlite
		`INSERT INTO archived_features (id,display_name,technical_name,description,created_at,updated_at) VALUES (?,?,?,?,?,?)`,
		f.ID, f.DisplayName, f.TechnicalName, f.Description, f.CreatedAt, f.UpdatedAt,
	)
	return err
}
