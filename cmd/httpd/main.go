package main

import (
	"context"
	"database/sql"
	"flag"
	"net/http"
	"os"
	"os/signal"
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

	r := chi.NewRouter()

	r.Use(
		hlog.NewHandler(log.Logger),
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).
				Info().
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("ACCESS")
		}),
	)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/features", func(r chi.Router) {
			r.Post("/", featureHandler.SaveFeature)
			r.Post("/request", nil) // Couldn't come up with a better name.

			r.Route("/{featureId}", func(r chi.Router) {
				r.Put("/", featureHandler.UpdateFeature)
				r.Post("/customers", featureHandler.SaveFeatureCustomers)
			})
		})

		r.Route("/archived_features", func(r chi.Router) {
			r.Post("/", featureHandler.SaveArchivedFeature)
		})
	})

	server := http.Server{
		Addr:         config.ServerAddr(),
		Handler:      r,
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
