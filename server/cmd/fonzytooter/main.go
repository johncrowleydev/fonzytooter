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

	"github.com/johncrowleydev/fonzytooter/server/internal/auth"
	"github.com/johncrowleydev/fonzytooter/server/internal/config"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
	"github.com/johncrowleydev/fonzytooter/server/internal/httpapi"
	"github.com/johncrowleydev/fonzytooter/server/internal/learner"
	"github.com/johncrowleydev/fonzytooter/server/internal/review"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
	openrouterprovider "github.com/johncrowleydev/fonzytooter/server/internal/tutor/openrouter"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutorlearning"
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
	authService := auth.NewService(db, auth.SessionConfig{
		Secure: cfg.Authentication.SecureCookie,
		TTL:    cfg.Authentication.SessionTTL,
	})
	if err := authService.ProvisionBootstrap(ctx, auth.BootstrapConfig{
		Username:    cfg.Authentication.Username,
		Password:    cfg.Authentication.Password,
		DisplayName: cfg.Authentication.DisplayName,
	}); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("configure authentication: %w", err)
	}
	tutorProvider, err := configuredTutorProvider(cfg)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	learnerService := learner.NewService(db, catalog)
	reviewService := review.NewService(db, catalog, review.SystemClock{})
	tutorService, err := configuredTutorService(tutorProvider, conversationStore, catalog, learnerService, reviewService)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	server := httpapi.NewServer(cfg.Address, tutorService, catalog, learnerService, reviewService, authService)
	return server, db, nil
}

func configuredTutorService(provider tutor.Provider, conversations *tutor.ConversationStore, catalog *curriculum.Catalog, learnerService *learner.Service, reviewService *review.Service) (*tutor.Service, error) {
	builder, err := tutorlearning.NewContextBuilder(catalog, learnerService)
	if err != nil {
		return nil, fmt.Errorf("configure tutor context: %w", err)
	}
	learningTools, err := tutorlearning.NewTools(tutorlearning.Services{Catalog: catalog, Learner: learnerService, Review: reviewService})
	if err != nil {
		return nil, fmt.Errorf("configure tutor learning tools: %w", err)
	}
	registry, err := tutor.NewToolRegistry(learningTools...)
	if err != nil {
		return nil, fmt.Errorf("configure tutor tool registry: %w", err)
	}
	manager, err := tutor.NewContextManager(
		conversations,
		tutor.ConservativeTokenEstimator{},
		tutor.ModelCompactor{Provider: provider, Fallback: tutor.RuleBasedCompactor{}},
		tutor.DefaultContextManagerConfig(),
	)
	if err != nil {
		return nil, fmt.Errorf("configure tutor context manager: %w", err)
	}
	service, err := tutor.NewRuntimeService(tutor.RuntimeConfig{
		Provider: provider, Store: conversations, Tools: registry, ContextManager: manager,
		ContextBuilder: builder, MaxModelRounds: tutor.DefaultMaxModelRounds,
	})
	if err != nil {
		return nil, fmt.Errorf("configure tutor runtime: %w", err)
	}
	return service, nil
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
