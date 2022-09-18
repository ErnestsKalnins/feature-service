package feature

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	_ "github.com/mattn/go-sqlite3"

	"feature/pkg/config"
)

func TestSaveFeature(t *testing.T) {
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
		expiryDate    = time.Now().Truncate(time.Second)
		expiryDateUTC = expiryDate.UTC()
	)

	tests := map[string]struct {
		features []feature
		timeFunc func() time.Time
		uuidFunc func() (uuid.UUID, error)

		body string

		wantStatus   int
		wantBody     string
		wantFeatures []feature
	}{
		"successfully persist the feature": {
			timeFunc: func() time.Time { return refTime },
			uuidFunc: func() (uuid.UUID, error) { return generatedUUID, nil },

			body: `{"displayName":"My Feature 1","technicalName":"my-feature-1","expiresOn":"` + expiryDate.Format(time.RFC3339) + `","description":"Placeholder text for feature description."}`,

			wantStatus: http.StatusCreated,
			wantFeatures: []feature{{
				ID:            generatedUUID,
				DisplayName:   ptr("My Feature 1"),
				TechnicalName: "my-feature-1",
				ExpiresOn:     &expiryDateUTC,
				Description:   ptr("Placeholder text for feature description."),
				Inverted:      false,
				CreatedAt:     refTime.UTC(),
				UpdatedAt:     refTime.UTC(),
			}},
		},
		"feature with the same technical name already exists": {
			timeFunc: func() time.Time { return refTime },
			features: []feature{{
				ID:            existingUUID,
				TechnicalName: "my-feature-1",
			}},
			uuidFunc: func() (uuid.UUID, error) { return generatedUUID, nil },

			body: `{"displayName":"My Feature 1","technicalName":"my-feature-1","expiresOn":"` + expiryDate.Format(time.RFC3339) + `","description":"Placeholder text for feature description."}`,

			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"save feature: UNIQUE constraint failed: features.technical_name"}`,
			wantFeatures: []feature{{
				ID:            existingUUID,
				TechnicalName: "my-feature-1",
			}},
		},
		"failed to generate feature ID": {
			uuidFunc: func() (uuid.UUID, error) { return uuid.Nil, errors.New("test error") },

			body: `{"technicalName":"my-feature-1"}`,

			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"generate feature id: test error"}`,
		},
		"request body contains invalid field values": {
			body: `{}`,

			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"validate feature: 'technicalName' must be at least 5 characters long"}`,
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
			service.uuidFunc = test.uuidFunc
			handler := NewHandler(service)

			r := chi.NewRouter()
			r.Post("/features", handler.SaveFeature)

			req := httptest.NewRequest(
				http.MethodPost,
				"/features",
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

func (s Store) findAllFeatures(ctx context.Context) ([]feature, error) {
	rs, err := s.db.QueryContext(
		ctx,
		//language=sqlite
		`SELECT id,display_name,technical_name,expires_on,description,inverted,created_at,updated_at FROM features`,
	)
	if err != nil {
		return nil, err
	}

	var fs []feature
	for rs.Next() {
		var f feature
		if err := rs.Scan(
			&f.ID,
			&f.DisplayName,
			&f.TechnicalName,
			&f.ExpiresOn,
			&f.Description,
			&f.Inverted,
			&f.CreatedAt,
			&f.UpdatedAt,
		); err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}

	if err := rs.Err(); err != nil {
		return nil, err
	}

	if err := rs.Close(); err != nil {
		return nil, err
	}

	return fs, nil
}

func assertFeatures(t *testing.T, store Store, want ...feature) {
	t.Helper()
	got, err := store.findAllFeatures(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Features not equal.\nwant: %v\ngot:  %v", want, got)
	}
}

func ptr[T any](t T) *T { return &t }

func setupFeatures(t *testing.T, store Store, features ...feature) {
	t.Helper()
	for _, f := range features {
		if err := store.saveFeature(context.Background(), f); err != nil {
			t.Fatalf("failed to set up features table: %s\n", err)
		}
	}
}
