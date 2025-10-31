package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type LogEntry struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Source    string `json:"source"`
	Level     string `json:"level"`
}

type AnalysisRequest struct {
	Logs []LogEntry `json:"logs"`
}

type AnalysisResponse struct {
	Processed int `json:"processed"`
	Anomalies int `json:"anomalies"`
	Results   []struct {
		LogID       int     `json:"log_id"`
		AnomalyScore float64 `json:"anomaly_score"`
		IsAnomaly   bool    `json:"is_anomaly"`
	} `json:"results"`
}

type Config struct {
	WatchDir      string `json:"watch_dir"`
	MLServiceURL  string `json:"ml_service_url"`
	BatchSize     int    `json:"batch_size"`
	CheckInterval string `json:"check_interval"`
}

func main() {
	config := Config{
		WatchDir:      getEnv("WATCH_DIR", "./logs"),
		MLServiceURL:  getEnv("ML_SERVICE_URL", "http://localhost:8000"),
		BatchSize:     10,
		CheckInterval: getEnv("CHECK_INTERVAL", "5s"),
	}

	// Create watch directory
	if err := os.MkdirAll(config.WatchDir, 0755); err != nil {
		log.Fatalf("Failed to create watch directory: %v", err)
	}

	log.Printf("Starting AIOps Log Ingestor")
	log.Printf("Watching: %s", config.WatchDir)
	log.Printf("ML Service: %s", config.MLServiceURL)

	// Start file watcher
	if err := startWatcher(config); err != nil {
		log.Fatal(err)
	}
}

func startWatcher(config Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(config.WatchDir)
	if err != nil {
		return err
	}

	// Process existing files first
	if err := processExistingFiles(config); err != nil {
		log.Printf("Warning: failed to process existing files: %v", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Printf("File modified: %s", event.Name)
				go processFile(event.Name, config)
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				log.Printf("File created: %s", event.Name)
				go processFile(event.Name, config)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func processExistingFiles(config Config) error {
	files, err := filepath.Glob(filepath.Join(config.WatchDir, "*.log"))
	if err != nil {
		return err
	}

	for _, file := range files {
		log.Printf("Processing existing file: %s", file)
		go processFile(file, config)
	}

	return nil
}

func processFile(filePath string, config Config) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v", filePath, err)
		return
	}

	// Simple line-by-line processing
	var logs []LogEntry
	// In real implementation, parse log lines properly
	logEntry := LogEntry{
		Message:   string(content),
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    filePath,
		Level:     "INFO",
	}
	logs = append(logs, logEntry)

	// Send for analysis
	if err := sendForAnalysis(logs, config); err != nil {
		log.Printf("Error sending logs for analysis: %v", err)
	}
}

func sendForAnalysis(logs []LogEntry, config Config) error {
	request := AnalysisRequest{Logs: logs}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := http.Post(config.MLServiceURL+"/analyze", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result AnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	log.Printf("Analysis complete: processed=%d, anomalies=%d", result.Processed, result.Anomalies)
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
