package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
	"github.com/johncrowleydev/helix-academy/server/internal/httpapi"
	"github.com/johncrowleydev/helix-academy/server/internal/tutor"
)

func main() {
	application := httpapi.NewAPI(
		tutor.NewService(tutor.NewUnavailableProvider()),
		curriculum.NewEmptyCatalog(),
		nil,
		nil,
	)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(application.Spec); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "write OpenAPI: %v\n", err)
		os.Exit(1)
	}
}
