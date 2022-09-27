package feature

import (
	"context"
	"database/sql"
	"feature/pkg/render"
	"feature/pkg/set"
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

	tx, commit, rollback, err := svc.store.beginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelDefault,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer rollback()

	if err := tx.saveFeature(ctx, f); err != nil {
		return fmt.Errorf("save feature: %w", err)
	}

	var cs []customer
	for _, cid := range f.CustomerIDs {
		id, err := svc.uuidFunc()
		if err != nil {
			return fmt.Errorf("generate customer feature join table id: %w", err)
		}
		cs = append(cs, customer{
			ID:         id,
			FeatureID:  f.ID,
			CustomerID: cid,
		})
	}

	if err := tx.saveCustomers(ctx, cs...); err != nil {
		return fmt.Errorf("save customers: %w", err)
	}

	if err := commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (svc Service) updateFeature(ctx context.Context, lastUpdatedAt time.Time, f feature) error {
	if err := f.validate(); err != nil {
		return fmt.Errorf("validate feature: %w", err)
	}

	tx, commit, rollback, err := svc.store.beginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer rollback()

	f.UpdatedAt = svc.timeFunc()
	if err := tx.updateFeature(ctx, lastUpdatedAt, f); err != nil {
		return fmt.Errorf("update feature: %w", err)
	}

	ids, err := tx.findCustomerIDsByFeatureID(ctx, f.ID)
	if err != nil {
		return fmt.Errorf("find customer ids by feature id: %w", err)
	}

	var (
		newIDs     = set.Of(f.CustomerIDs...)
		currentIDs = set.Of(ids...)
		common     = set.Intersection(newIDs, currentIDs)
		toSave     = set.Sub(newIDs, common)
		toDelete   = set.Sub(currentIDs, common)
	)

	var newCustomers []customer
	for _, s := range toSave.ToSlice() {
		id, err := svc.uuidFunc()
		if err != nil {
			return fmt.Errorf("generate customer feature join table id: %w", err)
		}
		newCustomers = append(newCustomers, customer{
			ID:         id,
			FeatureID:  f.ID,
			CustomerID: s,
		})
	}

	if err := tx.saveCustomers(ctx, newCustomers...); err != nil {
		return fmt.Errorf("save new customers: %w", err)
	}

	if err := tx.deleteCustomersByCustomerIDs(ctx, toDelete.ToSlice()...); err != nil {
		return fmt.Errorf("delete removed customers: %w", err)
	}

	if err := commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
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
