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

func TestSaveArchivedFeature(t *testing.T) {
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
		refTime    = time.Now().Truncate(time.Second)
		expiryDate = time.Now().Truncate(time.Second)
		//expiryDateUTC = expiryDate.UTC()
	)

	tests := map[string]struct {
		features []feature
		timeFunc func() time.Time

		body string

		wantStatus           int
		wantBody             string
		wantArchivedFeatures []archivedFeature
	}{
		"successfully archive a feature": {
			features: []feature{{
				ID:            existingUUID,
				DisplayName:   ptr("My Feature 1"),
				TechnicalName: "my-feature-1",
				ExpiresOn:     &expiryDate,
				Description:   ptr("My Feature 1 description."),
				Inverted:      false,
				CreatedAt:     refTime.AddDate(0, 0, -2),
				UpdatedAt:     refTime.AddDate(0, 0, -1),
			}},
			timeFunc: func() time.Time { return refTime },

			body: `{"featureId":"` + existingUUID.String() + `"}`,

			wantStatus: http.StatusCreated,
			wantArchivedFeatures: []archivedFeature{{
				ID:            existingUUID,
				DisplayName:   ptr("My Feature 1"),
				TechnicalName: "my-feature-1",
				Description:   ptr("My Feature 1 description."),
				CreatedAt:     refTime.UTC(),
				UpdatedAt:     refTime.UTC(),
			}},
		},
		"request body refers to non-existing feature": {
			body: `{"featureId":"` + existingUUID.String() + `"}`,

			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"find feature: feature bb7fe5b6-24a5-4218-bc61-b487bbad9580 does not exist"}`,
		},
		"request body contains unknown fields": {
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

			service := NewService(*tx)
			service.timeFunc = test.timeFunc
			handler := NewHandler(service)

			r := chi.NewRouter()
			r.Post("/archived_features", handler.SaveArchivedFeature)

			req := httptest.NewRequest(
				http.MethodPost,
				"/archived_features",
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

			assertArchivedFeatures(t, *tx, test.wantArchivedFeatures...)
		})
	}
}

func (s Store) findAllArchivedFeatures(ctx context.Context) ([]archivedFeature, error) {
	rs, err := s.db.QueryContext(
		ctx,
		//language=sqlite
		`SELECT id,display_name,technical_name,description,created_at,updated_at FROM archived_features`,
	)
	if err != nil {
		return nil, err
	}

	var afs []archivedFeature
	for rs.Next() {
		var af archivedFeature
		if err := rs.Scan(
			&af.ID,
			&af.DisplayName,
			&af.TechnicalName,
			&af.Description,
			&af.CreatedAt,
			&af.UpdatedAt,
		); err != nil {
			return nil, err
		}
		afs = append(afs, af)
	}

	if err := rs.Err(); err != nil {
		return nil, err
	}

	if err := rs.Close(); err != nil {
		return nil, err
	}
	return afs, nil
}

func assertArchivedFeatures(t *testing.T, store Store, want ...archivedFeature) {
	t.Helper()
	got, err := store.findAllArchivedFeatures(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Archived features not equal.\nwant: %v\ngot:  %v\n", want, got)
	}
}
