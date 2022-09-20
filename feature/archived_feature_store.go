package feature

import "context"

func (s Store) saveArchivedFeature(ctx context.Context, f feature) error {
	r := featureToRow(f)
	_, err := s.db.ExecContext(
		ctx,
		//language=sqlite
		`INSERT INTO archived_features (id,display_name,technical_name,description,created_at,updated_at) VALUES (?,?,?,?,?,?)`,
		r.ID, r.DisplayName, r.TechnicalName, r.Description, r.CreatedAt, r.UpdatedAt,
	)
	return err
}
