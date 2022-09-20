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
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSaveFeatureCustomers(t *testing.T) {
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
		generatedUUID = uuid.MustParse("44eeeacf-8d5d-4c68-bbe9-3e58c5ae6915")
		refTime       = time.Now().Truncate(time.Second)
	)

	tests := map[string]struct {
		features  []feature
		customers []customer
		uuidFunc  func() (uuid.UUID, error)

		featureID string
		body      string

		wantStatus    int
		wantBody      string
		wantCustomers []customer
	}{
		"successfully add a customer to a feature": {
			uuidFunc: func() (uuid.UUID, error) { return generatedUUID, nil },
			features: []feature{{
				ID:            existingUUID,
				TechnicalName: "my-feature-1",
				CreatedAt:     refTime,
				UpdatedAt:     refTime,
			}},

			featureID: existingUUID.String(),
			body:      `{"customerIds":["customer-1"]}`,

			wantStatus: http.StatusCreated,
			wantCustomers: []customer{{
				ID:         generatedUUID,
				FeatureID:  existingUUID,
				CustomerID: "customer-1",
			}},
		},
		"attempt to add a customer to a feature twice": {
			uuidFunc: func() (uuid.UUID, error) { return generatedUUID, nil },
			features: []feature{{
				ID:            existingUUID,
				TechnicalName: "my-feature-1",
				CreatedAt:     refTime,
				UpdatedAt:     refTime,
			}},
			customers: []customer{{
				ID:         existingUUID,
				FeatureID:  existingUUID,
				CustomerID: "customer-1",
			}},

			featureID: existingUUID.String(),
			body:      `{"customerIds":["customer-1"]}`,

			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"save customers: UNIQUE constraint failed: customer_features.customer_id, customer_features.feature_id"}`,
			wantCustomers: []customer{{
				ID:         existingUUID,
				FeatureID:  existingUUID,
				CustomerID: "customer-1",
			}},
		},
		"attempt to add customer to non-existing feature": {
			uuidFunc: func() (uuid.UUID, error) { return generatedUUID, nil },

			featureID: existingUUID.String(),
			body:      `{"customerIds":["customer-1"]}`,

			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"save customers: FOREIGN KEY constraint failed"}`,
		},
		"request body contains no customer ids": {
			featureID: existingUUID.String(),
			body:      `{}`,

			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"no customer IDs given"}`,
		},
		"request body contains unknown fields": {
			featureID: existingUUID.String(),
			body:      `{"foo":"bar"}`,

			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"decode request body: json: unknown field \"foo\""}`,
		},
		"missing request body": {
			featureID: existingUUID.String(),

			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"decode request body: EOF"}`,
		},
		"bad feature id": {
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"parse feature id: invalid UUID length: 0"}`,
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
			service.uuidFunc = test.uuidFunc
			handler := NewHandler(service)

			r := chi.NewRouter()
			r.Post("/features/{featureId}/customers", handler.SaveFeatureCustomers)

			req := httptest.NewRequest(
				http.MethodPost,
				"/features/"+test.featureID+"/customers",
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

			assertCustomers(t, *tx, test.wantCustomers...)
		})
	}
}

func (s Store) findAllCustomerFeatures(ctx context.Context) ([]customer, error) {
	rs, err := s.db.QueryContext(
		ctx,
		//language=sqlite
		`SELECT id,feature_id,customer_id FROM customer_features`,
	)
	if err != nil {
		return nil, err
	}

	var cs []customer
	for rs.Next() {
		var c customer
		if err := rs.Scan(&c.ID, &c.FeatureID, &c.CustomerID); err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}

	if err := rs.Err(); err != nil {
		return nil, err
	}

	if err := rs.Close(); err != nil {
		return nil, err
	}

	return cs, nil
}

func assertCustomers(t *testing.T, store Store, want ...customer) {
	t.Helper()
	got, err := store.findAllCustomerFeatures(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Customers not equal.\nwant: %v\ngot:  %v", want, got)
	}
}

func setupCustomers(t *testing.T, store Store, cs ...customer) {
	t.Helper()
	if err := store.saveCustomers(context.Background(), cs...); err != nil {
		t.Errorf("failed to set up customer_features table: %s", err)
	}
}
