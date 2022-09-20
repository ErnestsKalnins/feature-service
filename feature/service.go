package feature

import (
	"context"
	"database/sql"
	"feature/pkg/render"
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

	now := svc.timeFunc()
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

	f.UpdatedAt = svc.timeFunc()
	if err := svc.store.updateFeature(ctx, lastUpdatedAt, f); err != nil {
		return fmt.Errorf("update feature: %w", err)
	}

	return nil
}

func (svc Service) archiveFeature(ctx context.Context, featureID uuid.UUID) error {
	tx, commit, rollback, err := svc.store.beginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer rollback()

	f, err := tx.findFeature(ctx, featureID)
	if err != nil {
		return fmt.Errorf("find feature: %w", err)
	}

	now := svc.timeFunc()
	f.CreatedAt, f.UpdatedAt = now, now

	if err := tx.saveArchivedFeature(ctx, *f); err != nil {
		return fmt.Errorf("save archived feature: %w", err)
	}

	if err := tx.deleteFeature(ctx, featureID); err != nil {
		return fmt.Errorf("delete feature: %w", err)
	}

	if err := commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

var errNoCustomers = render.NewBadRequest("no customer IDs given")

func (svc Service) addCustomersToFeature(ctx context.Context, featureID uuid.UUID, customerIDs []string) error {
	if len(customerIDs) == 0 {
		return errNoCustomers
	}

	var customers []customer
	for _, customerID := range customerIDs {
		id, err := svc.uuidFunc()
		if err != nil {
			return fmt.Errorf("generate customer feature id: %w", err)
		}
		customers = append(customers, customer{
			ID:         id,
			FeatureID:  featureID,
			CustomerID: customerID,
		})
	}

	if err := svc.store.saveCustomers(ctx, customers...); err != nil {
		return fmt.Errorf("save customers: %w", err)
	}

	return nil
}

var errNoFeatureNames = render.NewBadRequest("no feature technical names given")

func (svc Service) findCustomerFeaturesByTechnicalNames(ctx context.Context, customerID string, technicalNames ...string) ([]customerFeature, error) {
	if len(technicalNames) == 0 {
		return nil, errNoFeatureNames
	}

	cfs, err := svc.store.findCustomerFeaturesByTechnicalNames(ctx, customerID, svc.timeFunc(), technicalNames...)
	if err != nil {
		return nil, fmt.Errorf("find customer features by technical names: %w", err)
	}
	return cfs, nil
}
