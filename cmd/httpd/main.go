package main

import (
	"context"
	"database/sql"
	"embed"
	"flag"
	"github.com/go-chi/cors"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"feature/feature"
	"feature/pkg/config"
)

//go:embed dist
var app embed.FS

func main() {
	envFile := flag.String("env-file", "", "Path to env file containing configuration.")

	flag.Parse()

	if *envFile != "" {
		viper.SetConfigFile(*envFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().
				Err(err).
				Msg("failed to read configuration")
		}
	}

	db, err := sql.Open("sqlite3", config.DSN())
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to open connection to database")
	}

	if err := db.Ping(); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to ping database")
	}

	featureStore := feature.NewStore(db)
	featureService := feature.NewService(featureStore)
	featureHandler := feature.NewHandler(featureService)

	apiHandler := chi.NewRouter()

	apiHandler.Route("/api/v1", func(r chi.Router) {
		r.Route("/features", func(r chi.Router) {
			r.Get("/", featureHandler.ListFeatures)
			r.Post("/", featureHandler.SaveFeature)
			r.Post("/request", featureHandler.RequestFeaturesAsCustomer) // Couldn't come up with a better name.

			r.Route("/{featureId}", func(r chi.Router) {
				r.Get("/", featureHandler.GetFeature)
				r.Put("/", featureHandler.UpdateFeature)
				r.Post("/customers", featureHandler.SaveFeatureCustomers)
			})
		})

		r.Route("/archived_features", func(r chi.Router) {
			r.Post("/", featureHandler.SaveArchivedFeature)
		})
	})

	appDir, err := fs.Sub(app, "dist/frontend")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to open app filesystem")
	}

	appHandler := http.FileServer(http.FS(appDir))

	rootHandler := chi.NewRouter()

	rootHandler.Use(
		cors.Handler(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		}),
		hlog.NewHandler(log.Logger),
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).
				Info().
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Stringer("url", r.URL).
				Msg("ACCESS")
		}),
	)

	rootHandler.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api"):
			apiHandler.ServeHTTP(w, r)
		case strings.HasPrefix(r.URL.Path, "/features"):
			r.URL.Path = "" // This is done so client-side routing can take over.
			fallthrough
		default:
			appHandler.ServeHTTP(w, r)
		}
	})

	server := http.Server{
		Addr:         config.ServerAddr(),
		Handler:      rootHandler,
		ReadTimeout:  config.ServerReadTimeout(),
		WriteTimeout: config.ServerWriteTimeout(),
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM)
		<-sigint

		if err := server.Shutdown(context.Background()); err != nil {
			log.Error().
				Err(err).
				Msg("error shutting down HTTP server")
		}
		close(idleConnsClosed)
	}()

	log.Info().
		Msg("serving the application over HTTP")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().
			Err(err).
			Msg("error serving the application")
	}
	<-idleConnsClosed
}
