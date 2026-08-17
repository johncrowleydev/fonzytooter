package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

type Server struct {
	tutor *tutor.Service
}

func NewServer(address string, tutorService *tutor.Service) *http.Server {
	api := &Server{tutor: tutorService}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", api.handleHealth)
	mux.HandleFunc("POST /api/tutor/turn", api.handleTutorTurn)

	return &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleTutorTurn(w http.ResponseWriter, r *http.Request) {
	var request tutor.TurnRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(request.Message) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message is required"})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming is unavailable"})
		return
	}

	events, err := s.tutor.StreamTurn(r.Context(), request)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			continue
		}

		_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
		flusher.Flush()
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
