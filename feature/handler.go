package feature

import (
	"encoding/json"
	"feature/pkg/slices"
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

// ListFeatures renders all features to the client.
func (h Handler) ListFeatures(w http.ResponseWriter, r *http.Request) {
	fs, err := h.service.store.findAllFeatures(r.Context())
	if err != nil {
		hlog.FromRequest(r).
			Error().
			Err(err).
			Msg("failed to find all features")
		render.Error(w, err)
		return
	}

	render.JSON(w, slices.Map(responseFromFeature, fs...))
}

func (h Handler) GetFeature(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "featureId"))
	if err != nil {
		render.Error(w, render.NewBadRequest(fmt.Sprintf("parse feature id: %s", err)))
		return
	}

	f, err := h.service.store.findFeatureWithClients(r.Context(), id)
	if err != nil {
		hlog.FromRequest(r).
			Error().
			Err(err).
			Msg("failed to find feature")
		render.Error(w, err)
		return
	}

	render.JSON(w, responseFromFeature(*f))
}

func responseFromFeature(f feature) featureResponse {
	res := featureResponse{
		ID:            f.ID,
		DisplayName:   f.DisplayName,
		TechnicalName: f.TechnicalName,
		Description:   f.Description,
		Inverted:      f.Inverted,
		CreatedAt:     f.CreatedAt.UnixMilli(),
		UpdatedAt:     f.UpdatedAt.UnixMilli(),
	}
	if f.ExpiresOn != nil {
		res.ExpiresOn = new(int64)
		*res.ExpiresOn = f.ExpiresOn.UnixMilli()
	}
	if 0 < len(f.CustomerIDs) {
		res.CustomerIDs = f.CustomerIDs
	}
	return res
}

type featureResponse struct {
	ID            uuid.UUID `json:"id"`
	DisplayName   *string   `json:"displayName,omitempty"`
	TechnicalName string    `json:"technicalName"`
	ExpiresOn     *int64    `json:"expiresOn,omitempty"`
	Description   *string   `json:"description,omitempty"`
	Inverted      bool      `json:"inverted"`
	CreatedAt     int64     `json:"createdAt"`
	UpdatedAt     int64     `json:"updatedAt"`
	CustomerIDs   []string  `json:"customerIds,omitempty"`
}

type saveFeatureRequest struct {
	DisplayName   *string  `json:"displayName"`
	TechnicalName string   `json:"technicalName"`
	ExpiresOn     *int64   `json:"expiresOn"`
	Description   *string  `json:"description"`
	Inverted      bool     `json:"inverted"`
	CustomerIDs   []string `json:"customerIds"`
}

func (r saveFeatureRequest) toFeature() feature {
	res := feature{
		DisplayName:   r.DisplayName,
		TechnicalName: r.TechnicalName,
		Description:   r.Description,
		Inverted:      r.Inverted,
		CustomerIDs:   r.CustomerIDs,
	}
	if r.ExpiresOn != nil {
		res.ExpiresOn = new(time.Time)
		*res.ExpiresOn = time.UnixMilli(*r.ExpiresOn)
	}
	return res
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
	LastUpdatedAt int64              `json:"lastUpdatedAt"`
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
	if err := h.service.updateFeature(r.Context(), time.UnixMilli(req.LastUpdatedAt), f); err != nil {
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

type featureRequest struct {
	Request struct {
		CustomerID string `json:"customerId"`
		Features   []struct {
			Name string `json:"name"`
		} `json:"features"`
	} `json:"featureRequest"`
}

func (r featureRequest) featureTechnicalNames() []string {
	res := make([]string, len(r.Request.Features))
	for i := range r.Request.Features {
		res[i] = r.Request.Features[i].Name
	}
	return res
}

func (h Handler) RequestFeaturesAsCustomer(w http.ResponseWriter, r *http.Request) {
	var req featureRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		render.Error(w, render.NewBadRequest(fmt.Sprintf("decode request body: %s", err)))
		return
	}

	cfs, err := h.service.findCustomerFeaturesByTechnicalNames(r.Context(), req.Request.CustomerID, req.featureTechnicalNames()...)
	if err != nil {
		hlog.FromRequest(r).
			Error().
			Err(err).
			Msg("failed to retrieve features by technical names")
		render.Error(w, err)
		return
	}

	render.JSON(w, responseFromCustomerFeatures(cfs))
}

func responseFromCustomerFeatures(cfs []customerFeature) customerFeaturesResponse {
	features := make([]customerFeatureResponse, len(cfs))
	for i, cf := range cfs {
		features[i] = customerFeatureResponse{
			Name:     cf.TechnicalName,
			Active:   cf.isActive(),
			Inverted: cf.Inverted,
			Expired:  cf.Expired,
		}
	}
	return customerFeaturesResponse{
		Features: features,
	}
}

type customerFeaturesResponse struct {
	Features []customerFeatureResponse `json:"features"`
}

type customerFeatureResponse struct {
	Name     string `json:"name"`
	Active   bool   `json:"active"`
	Inverted bool   `json:"inverted"`
	Expired  bool   `json:"expired"`
}
