package main

import (
	"database/sql"
	"feature/pkg/config"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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

	err = filepath.Walk("migrations", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}

		b, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		if _, err := db.Exec(string(b)); err != nil {
			return fmt.Errorf("execute query: %w", err)
		}

		return nil
	})
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to execute migrations")
	}
}
