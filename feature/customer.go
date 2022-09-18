package feature

import "github.com/google/uuid"

type customer struct {
	ID         uuid.UUID
	FeatureID  uuid.UUID
	CustomerID string
}
