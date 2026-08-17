package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/johncrowleydev/fonzytooter/server/internal/config"
	"github.com/johncrowleydev/fonzytooter/server/internal/httpapi"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

func main() {
	cfg := config.FromEnv()

	tutorService := tutor.NewService(tutor.NewUnavailableProvider())
	server := httpapi.NewServer(cfg.Address, tutorService)

	log.Printf("fonzytooter API listening on %s", cfg.Address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
