package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/johncrowleydev/fonzytooter/server/internal/config"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/httpapi"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

func main() {
	cfg := config.FromEnv()
	if info, err := os.Stat(cfg.CurriculumPath); err != nil {
		log.Fatalf("open curriculum path %q: %v", cfg.CurriculumPath, err)
	} else if !info.IsDir() {
		log.Fatalf("open curriculum path %q: not a directory", cfg.CurriculumPath)
	}
	catalog, err := curriculum.Load(os.DirFS(cfg.CurriculumPath))
	if err != nil {
		log.Fatalf("load curriculum from %q: %v", cfg.CurriculumPath, err)
	}

	tutorService := tutor.NewService(tutor.NewUnavailableProvider())
	server := httpapi.NewServer(cfg.Address, tutorService, catalog)

	log.Printf("fonzytooter API listening on %s", cfg.Address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
