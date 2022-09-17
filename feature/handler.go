package feature

import (
	"encoding/json"
	"github.com/rs/zerolog/hlog"
	"net/http"
	"time"

	"feature/pkg/render"
)

// NewHandler initializes and returns a new Handler.
func NewHandler(service Service) Handler {
	return Handler{service: service}
}

// Handler exposes Service methods over HTTP.
type Handler struct {
	service Service
}

type saveFeatureRequest struct {
	DisplayName   *string    `json:"displayName"`
	TechnicalName string     `json:"technicalName"`
	ExpiresOn     *time.Time `json:"expiresOn"`
	Description   *string    `json:"description"`
}

func (r saveFeatureRequest) toFeature() feature {
	return feature{
		DisplayName:   r.DisplayName,
		TechnicalName: r.TechnicalName,
		ExpiresOn:     r.ExpiresOn,
		Description:   r.Description,
	}
}

// SaveFeature persists the feature received via JSON request body.
func (h Handler) SaveFeature(w http.ResponseWriter, r *http.Request) {
	var req saveFeatureRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		render.Error(w, render.TagBadRequest(err))
		return
	}

	if err := h.service.saveFeature(r.Context(), req.toFeature()); err != nil {
		hlog.FromRequest(r).
			Error().
			Err(err).
			Msg("failed to save feature")
		render.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
