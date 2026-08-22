package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/config"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
	"github.com/johncrowleydev/fonzytooter/server/internal/httpapi"
	"github.com/johncrowleydev/fonzytooter/server/internal/learner"
	"github.com/johncrowleydev/fonzytooter/server/internal/review"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
	openrouterprovider "github.com/johncrowleydev/fonzytooter/server/internal/tutor/openrouter"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, config.FromEnv(), database.Open); err != nil {
		log.Fatal(err)
	}
}

type databaseOpener func(context.Context, string) (*sql.DB, error)

func run(ctx context.Context, cfg config.Config, openDatabase databaseOpener) error {
	server, db, err := prepareServer(ctx, cfg, openDatabase)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close learner database: %v", err)
		}
	}()

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("fonzytooter API listening on %s", cfg.Address)
		serveErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			return fmt.Errorf("shut down HTTP server: %w", err)
		}
		if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}
}

func prepareServer(ctx context.Context, cfg config.Config, openDatabase databaseOpener) (*http.Server, *sql.DB, error) {
	db, err := openDatabase(ctx, cfg.DatabasePath)
	if err != nil {
		return nil, nil, fmt.Errorf("prepare learner database %q: %w", cfg.DatabasePath, err)
	}

	if info, err := os.Stat(cfg.CurriculumPath); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("open curriculum path %q: %w", cfg.CurriculumPath, err)
	} else if !info.IsDir() {
		_ = db.Close()
		return nil, nil, fmt.Errorf("open curriculum path %q: not a directory", cfg.CurriculumPath)
	}
	catalog, err := curriculum.Load(os.DirFS(cfg.CurriculumPath))
	if err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("load curriculum from %q: %w", cfg.CurriculumPath, err)
	}

	conversationStore := tutor.NewConversationStore(db)
	tutorProvider, err := configuredTutorProvider(cfg)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	tutorService := tutor.NewPersistentService(tutorProvider, conversationStore)
	learnerService := learner.NewService(db, catalog)
	reviewService := review.NewService(db, catalog, review.SystemClock{})
	server := httpapi.NewServer(cfg.Address, tutorService, catalog, learnerService, reviewService)
	return server, db, nil
}

func configuredTutorProvider(cfg config.Config) (tutor.Provider, error) {
	if !cfg.OpenRouter.Configured() {
		return tutor.NewUnavailableProvider(), nil
	}
	provider, err := openrouterprovider.New(openrouterprovider.Config{
		APIKey:  cfg.OpenRouter.APIKey,
		Model:   cfg.OpenRouter.Model,
		BaseURL: cfg.OpenRouter.BaseURL,
		Client:  &http.Client{Timeout: cfg.OpenRouter.Timeout},
	})
	if err != nil {
		return nil, fmt.Errorf("configure OpenRouter tutor provider: %w", err)
	}
	return provider, nil
}
