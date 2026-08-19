package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/httpapi"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

func main() {
	application := httpapi.NewAPI(
		tutor.NewService(tutor.NewUnavailableProvider()),
		curriculum.NewEmptyCatalog(),
	)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(application.Spec); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "write OpenAPI: %v\n", err)
		os.Exit(1)
	}
}
