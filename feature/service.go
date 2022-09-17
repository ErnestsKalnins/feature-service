package feature

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"time"
)

// NewService initializes and returns a new Service.
func NewService(store Store) Service {
	return Service{
		store:    store,
		timeFunc: time.Now,
		uuidFunc: uuid.NewRandom,
	}
}

// Service exposes business functionality related to feature toggling and
// querying.
type Service struct {
	store Store

	timeFunc func() time.Time
	uuidFunc func() (uuid.UUID, error)
}

func (svc Service) saveFeature(ctx context.Context, f feature) error {
	if err := f.validate(); err != nil {
		return fmt.Errorf("validate feature: %w", err)
	}

	id, err := svc.uuidFunc()
	if err != nil {
		return fmt.Errorf("generate feature id: %w", err)
	}
	f.ID = id

	if f.ExpiresOn != nil {
		*f.ExpiresOn = f.ExpiresOn.UTC()
	}

	now := svc.timeFunc().UTC()
	f.CreatedAt, f.UpdatedAt = now, now

	if err := svc.store.saveFeature(ctx, f); err != nil {
		return fmt.Errorf("save feature: %w", err)
	}

	return nil
}

func (svc Service) updateFeature(ctx context.Context, lastUpdatedAt time.Time, f feature) error {
	if err := f.validate(); err != nil {
		return fmt.Errorf("validate feature: %w", err)
	}

	f.UpdatedAt = svc.timeFunc().UTC()

	if f.ExpiresOn != nil {
		*f.ExpiresOn = f.ExpiresOn.UTC()
	}

	if err := svc.store.updateFeature(ctx, lastUpdatedAt, f); err != nil {
		return fmt.Errorf("update feature: %w", err)
	}

	return nil
}
