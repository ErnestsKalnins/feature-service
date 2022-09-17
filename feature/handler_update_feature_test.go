package feature

import (
	"context"
	"database/sql"
	"feature/pkg/config"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUpdateFeature(t *testing.T) {
	viper.SetConfigFile("../test.env")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite3", config.DSN())
	if err != nil {
		t.Fatal(err)
	}

	store := NewStore(db)

	var (
		existingUUID  = uuid.MustParse("bb7fe5b6-24a5-4218-bc61-b487bbad9580")
		lastUpdatedAt = time.Now().AddDate(0, 0, -7).Truncate(time.Second)
		refTime       = time.Now().Truncate(time.Second)
		expiryDate    = time.Now().Truncate(time.Second)
		expiryDateUTC = expiryDate.UTC()
	)

	tests := map[string]struct {
		features []feature
		timeFunc func() time.Time

		featureId string
		body      string

		wantStatus   int
		wantBody     string
		wantFeatures []feature
	}{
		"successfully update feature": {
			features: []feature{{
				ID:            existingUUID,
				DisplayName:   ptr("Feature #1"),
				TechnicalName: "feature-1",
				ExpiresOn:     &expiryDateUTC,
				Description:   ptr("Lorem ipsum."),
				CreatedAt:     lastUpdatedAt.UTC(),
				UpdatedAt:     lastUpdatedAt.UTC(),
			}},
			timeFunc: func() time.Time { return refTime },

			featureId: existingUUID.String(),
			body:      `{"lastUpdatedAt":"` + lastUpdatedAt.Format(time.RFC3339) + `","feature":{"displayName":"My Feature 1","technicalName":"my-feature-1","expiresOn":"` + expiryDate.Format(time.RFC3339) + `","description":"Placeholder text for feature description."}}`,

			wantStatus: http.StatusNoContent,
			wantFeatures: []feature{{
				ID:            existingUUID,
				DisplayName:   ptr("My Feature 1"),
				TechnicalName: "my-feature-1",
				ExpiresOn:     &expiryDateUTC,
				Description:   ptr("Placeholder text for feature description."),
				CreatedAt:     lastUpdatedAt.UTC(),
				UpdatedAt:     refTime.UTC(),
			}},
		},
		"updated feature exists, but client is sending a stale update": {
			features: []feature{{
				ID:            existingUUID,
				DisplayName:   ptr("Feature #1"),
				TechnicalName: "feature-1",
				ExpiresOn:     &expiryDateUTC,
				Description:   ptr("Lorem ipsum."),
				CreatedAt:     lastUpdatedAt.UTC(),
				UpdatedAt:     lastUpdatedAt.AddDate(0, 0, 1).UTC(),
			}},
			timeFunc: func() time.Time { return refTime },

			featureId: existingUUID.String(),
			body:      `{"lastUpdatedAt":"` + lastUpdatedAt.Format(time.RFC3339) + `","feature":{"displayName":"My Feature 1","technicalName":"my-feature-1","expiresOn":"` + expiryDate.Format(time.RFC3339) + `","description":"Placeholder text for feature description."}}`,

			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"update feature: feature bb7fe5b6-24a5-4218-bc61-b487bbad9580 does not exist"}`,
			wantFeatures: []feature{{
				ID:            existingUUID,
				DisplayName:   ptr("Feature #1"),
				TechnicalName: "feature-1",
				ExpiresOn:     &expiryDateUTC,
				Description:   ptr("Lorem ipsum."),
				CreatedAt:     lastUpdatedAt.UTC(),
				UpdatedAt:     lastUpdatedAt.AddDate(0, 0, 1).UTC(),
			}},
		},
		"updated feature doesn't exist": {
			timeFunc: func() time.Time { return refTime },

			featureId: existingUUID.String(),
			body:      `{"lastUpdatedAt":"` + lastUpdatedAt.Format(time.RFC3339) + `","feature":{"displayName":"My Feature 1","technicalName":"my-feature-1","expiresOn":"` + expiryDate.Format(time.RFC3339) + `","description":"Placeholder text for feature description."}}`,

			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"update feature: feature bb7fe5b6-24a5-4218-bc61-b487bbad9580 does not exist"}`,
		},
		"invalid request body": {
			featureId: existingUUID.String(),
			body:      `{}`,

			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"validate feature: 'technicalName' must be at least 5 characters long"}`,
		},
		"request body contains unknown fields": {
			featureId: existingUUID.String(),
			body:      `{"foo":"bar"}`,

			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"decode request body: json: unknown field \"foo\""}`,
		},
		"missing request body": {
			featureId: existingUUID.String(),

			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"decode request body: EOF"}`,
		},
		"bad feature id": {
			featureId: "bad",

			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"parse feature id: invalid UUID length: 3"}`,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tx, err := store.beginTx(context.Background(), &sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			})
			if err != nil {
				t.Fatalf("failed to begin transaction: %s\n", err)
			}

			t.Cleanup(func() {
				if err := tx.rollbackTx(); err != nil {
					t.Errorf("failed to rollback the transaction: %s\n", err)
				}
			})

			setupFeatures(t, *tx, test.features...)

			service := NewService(*tx)
			service.timeFunc = test.timeFunc
			handler := NewHandler(service)

			r := chi.NewRouter()
			r.Put("/features/{featureId}", handler.UpdateFeature)

			req := httptest.NewRequest(
				http.MethodPut,
				"/features/"+test.featureId,
				strings.NewReader(test.body),
			)
			res := httptest.NewRecorder()

			r.ServeHTTP(res, req)

			if res.Code != test.wantStatus {
				t.Errorf("Status codes not equal.\nwant: %d\ngot:  %d", test.wantStatus, res.Code)
			}

			resBody := strings.TrimSpace(res.Body.String())
			if resBody != test.wantBody {
				t.Errorf("Response bodies not equal.\nwant: %s\ngot:  %s", test.wantBody, resBody)
			}

			assertFeatures(t, *tx, test.wantFeatures...)
		})
	}
}
