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

func TestRequestFeaturesAsCustomer(t *testing.T) {
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
		existingUUID = uuid.MustParse("bb7fe5b6-24a5-4218-bc61-b487bbad9580")
		//generatedUUID = uuid.MustParse("44eeeacf-8d5d-4c68-bbe9-3e58c5ae6915")
		refTime   = time.Now().Truncate(time.Second).UTC()
		oneDayAgo = time.Now().Truncate(time.Second).AddDate(0, 0, -1).UTC()
		//expiryDate    = time.Now().Truncate(time.Second).UTC()
	)

	tests := map[string]struct {
		features  []feature
		customers []customer
		timeFunc  func() time.Time

		body string

		wantStatus int
		wantBody   string
	}{
		"successfully return inverted, non-expired feature the customer has": {
			timeFunc: func() time.Time { return refTime },
			features: []feature{{
				ID:            existingUUID,
				TechnicalName: "feature-1",
				Inverted:      true,
				CreatedAt:     refTime,
				UpdatedAt:     refTime,
			}},
			customers: []customer{{
				ID:         existingUUID,
				FeatureID:  existingUUID,
				CustomerID: "1234",
			}},

			body: `{"featureRequest":{"customerId":"1234","features":[{"name":"feature-1"}]}}`,

			wantStatus: http.StatusOK,
			wantBody:   `{"features":[{"name":"feature-1","active":false,"inverted":true,"expired":false}]}`,
		},
		"successfully return non-inverted, non-expired feature the customer has": {
			timeFunc: func() time.Time { return refTime },
			features: []feature{{
				ID:            existingUUID,
				TechnicalName: "feature-1",
				CreatedAt:     refTime,
				UpdatedAt:     refTime,
			}},
			customers: []customer{{
				ID:         existingUUID,
				FeatureID:  existingUUID,
				CustomerID: "1234",
			}},

			body: `{"featureRequest":{"customerId":"1234","features":[{"name":"feature-1"}]}}`,

			wantStatus: http.StatusOK,
			wantBody:   `{"features":[{"name":"feature-1","active":true,"inverted":false,"expired":false}]}`,
		},
		"successfully return non-inverted, expired feature the customer doesn't have": {
			timeFunc: func() time.Time { return refTime },
			features: []feature{{
				ID:            existingUUID,
				TechnicalName: "feature-1",
				ExpiresOn:     &oneDayAgo,
				CreatedAt:     refTime,
				UpdatedAt:     refTime,
			}},

			body: `{"featureRequest":{"customerId":"1234","features":[{"name":"feature-1"}]}}`,

			wantStatus: http.StatusOK,
			// customer is NOT in the list of the feature, but feature toggle expired:
			// {"name": "my-feature-d", "active": true, "inverted": false, "expired": true}
			// -----------------------------------^^^^
			// I assume this specification is false.
			wantBody: `{"features":[{"name":"feature-1","active":false,"inverted":false,"expired":true}]}`,
		},
		"successfully return inverted, non-expired feature the customer doesn't have": {
			timeFunc: func() time.Time { return refTime },
			features: []feature{{
				ID:            existingUUID,
				TechnicalName: "feature-1",
				Inverted:      true,
				CreatedAt:     refTime,
				UpdatedAt:     refTime,
			}},

			body: `{"featureRequest":{"customerId":"1234","features":[{"name":"feature-1"}]}}`,

			wantStatus: http.StatusOK,
			wantBody:   `{"features":[{"name":"feature-1","active":false,"inverted":true,"expired":false}]}`,
		},
		"successfully return non-inverted, non-expired feature the customer doesn't have": {
			timeFunc: func() time.Time { return refTime },
			features: []feature{{
				ID:            existingUUID,
				TechnicalName: "feature-1",
				Inverted:      false,
				CreatedAt:     refTime,
				UpdatedAt:     refTime,
			}},

			body: `{"featureRequest":{"customerId":"1234","features":[{"name":"feature-1"}]}}`,

			wantStatus: http.StatusOK,
			wantBody:   `{"features":[{"name":"feature-1","active":false,"inverted":false,"expired":false}]}`,
		},
		"requested feature doesn't exist": {
			timeFunc: func() time.Time { return refTime },

			body: `{"featureRequest":{"customerId":"1234","features":[{"name":"feature-1"}]}}`,

			wantStatus: http.StatusOK,
			wantBody:   `{"features":[]}`,
		},
		"no features requested": {
			body: `{"featureRequest":{"customerId":"1234"}}`,

			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"no feature technical names given"}`,
		},
		"request body has unknown fields": {
			body: `{"foo":"bar"}`,

			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"decode request body: json: unknown field \"foo\""}`,
		},
		"missing request body": {
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"decode request body: EOF"}`,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tx, _, rollback, err := store.beginTx(context.Background(), &sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			})
			if err != nil {
				t.Fatalf("failed to begin transaction: %s\n", err)
			}

			t.Cleanup(func() {
				if err := rollback(); err != nil {
					t.Errorf("failed to rollback the transaction: %s\n", err)
				}
			})

			setupFeatures(t, *tx, test.features...)
			setupCustomers(t, *tx, test.customers...)

			service := NewService(*tx)
			service.timeFunc = test.timeFunc
			handler := NewHandler(service)

			r := chi.NewRouter()
			r.Post("/", handler.RequestFeaturesAsCustomer)

			req := httptest.NewRequest(
				http.MethodPost,
				"/",
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
		})
	}
}
