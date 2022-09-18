package feature

import (
	"time"

	"github.com/google/uuid"
)

type archivedFeature struct {
	ID            uuid.UUID
	DisplayName   *string
	TechnicalName string
	Description   *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
