package feature

import "github.com/google/uuid"

type customer struct {
	ID         uuid.UUID
	FeatureID  uuid.UUID
	CustomerID string
}

type customerFeature struct {
	TechnicalName string
	Inverted      bool
	Expired       bool
	HasFeature    bool
}

func (cf customerFeature) isActive() bool {
	return cf.HasFeature && !cf.Inverted
}
