package domain

import "time"

type LogEntry struct {
	ID        string         `db:"id"`
	Service   string         `json:"service"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Hostname  string         `json:"hostname"`
	Env       string         `json:"env"`
	Timestamp time.Time      `json:"timestamp"`
	Metadata  map[string]any `json:metadata`
	Processed bool           `json:"processed"`
}
