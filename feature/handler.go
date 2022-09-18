package feature

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
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
	Inverted      bool       `json:"inverted"`
}

func (r saveFeatureRequest) toFeature() feature {
	return feature{
		DisplayName:   r.DisplayName,
		TechnicalName: r.TechnicalName,
		ExpiresOn:     r.ExpiresOn,
		Description:   r.Description,
		Inverted:      r.Inverted,
	}
}

// SaveFeature persists the feature received via JSON request body.
func (h Handler) SaveFeature(w http.ResponseWriter, r *http.Request) {
	var req saveFeatureRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		render.Error(w, render.NewBadRequest(fmt.Sprintf("decode request body: %s", err)))
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

type updateFeatureRequest struct {
	LastUpdatedAt time.Time          `json:"lastUpdatedAt"`
	Feature       saveFeatureRequest `json:"feature"`
}

// UpdateFeature updates an existing feature.
func (h Handler) UpdateFeature(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "featureId"))
	if err != nil {
		render.Error(w, render.NewBadRequest(fmt.Sprintf("parse feature id: %s", err)))
		return
	}

	var req updateFeatureRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		render.Error(w, render.NewBadRequest(fmt.Sprintf("decode request body: %s", err)))
		return
	}

	f := req.Feature.toFeature()
	f.ID = id
	if err := h.service.updateFeature(r.Context(), req.LastUpdatedAt, f); err != nil {
		hlog.FromRequest(r).
			Error().
			Err(err).
			Msg("failed to update feature")
		render.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type createArchivedFeatureRequest struct {
	FeatureID uuid.UUID `json:"featureId"`
}

// SaveArchivedFeature archives an existing feature.
func (h Handler) SaveArchivedFeature(w http.ResponseWriter, r *http.Request) {
	var req createArchivedFeatureRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		render.Error(w, render.NewBadRequest(fmt.Sprintf("decode request body: %s", err)))
		return
	}

	if err := h.service.archiveFeature(r.Context(), req.FeatureID); err != nil {
		hlog.FromRequest(r).
			Error().
			Err(err).
			Msg("failed to archive feature")
		render.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type saveFeatureCustomersRequest struct {
	CustomerIDs []string `json:"customerIds"`
}

// SaveFeatureCustomers persists the given customers to have access to the
// feature.
func (h Handler) SaveFeatureCustomers(w http.ResponseWriter, r *http.Request) {
	featureID, err := uuid.Parse(chi.URLParam(r, "featureId"))
	if err != nil {
		render.Error(w, render.NewBadRequest(fmt.Sprintf("parse feature id: %s", err)))
		return
	}

	var req saveFeatureCustomersRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		render.Error(w, render.NewBadRequest(fmt.Sprintf("decode request body: %s", err)))
		return
	}

	if err := h.service.addCustomersToFeature(r.Context(), featureID, req.CustomerIDs); err != nil {
		hlog.FromRequest(r).
			Error().
			Err(err).
			Msg("failed to add customers to feature")
		render.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
