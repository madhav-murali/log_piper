package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bluebox/internal/domain"
	"github.com/bluebox/internal/storage"

	"github.com/google/uuid"
)

type LogHandler struct {
	storage *storage.PostgresStorage
	// We can add NATS publisher here if needed
}

func NewLogHandler(storage *storage.PostgresStorage) *LogHandler {
	return &LogHandler{storage: storage}
}

func (h *LogHandler) IngestLog(w http.ResponseWriter, r *http.Request) {
	var log domain.LogEntry
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add metadata
	log.ID = uuid.New().String()
	log.Timestamp = time.Now().UTC()

	// TODO: Add more metadata, like from headers (e.g., service name might be in the header or in the body)

	if err := h.storage.InsertLog(r.Context(), &log); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(log)
}
