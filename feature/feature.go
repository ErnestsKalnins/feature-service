package feature

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// A feature toggle.
type feature struct {
	ID            uuid.UUID  `json:"id"`
	DisplayName   *string    `json:"displayName,omitempty"`
	TechnicalName string     `json:"technicalName"`
	ExpiresOn     *time.Time `json:"expiresOn,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Inverted      bool       `json:"inverted"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	CustomerIDs   []string   `json:"customerIds,omitempty"`
}

func (f feature) validate() error {
	var errs errFeatureInvalid

	if len(f.TechnicalName) < 5 {
		errs = append(errs, "'technicalName' must be at least 5 characters long")
	}

	if len(errs) != 0 {
		return errs
	}
	return nil
}

type errFeatureInvalid []string

func (e errFeatureInvalid) Error() string {
	return strings.Join(e, ", ")
}

func (e errFeatureInvalid) Code() int {
	return http.StatusBadRequest
}
