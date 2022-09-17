package feature

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// A feature toggle.
type feature struct {
	ID            uuid.UUID
	DisplayName   *string
	TechnicalName string
	ExpiresOn     *time.Time
	Description   *string
	Inverted      bool
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
