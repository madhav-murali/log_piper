package domain

import "time"

type MetricEntry struct {
	Id        string         `json:id`
	Service   string         `json:"service"`
	Metric    string         `json:"metric"`
	Value     float64        `json:"value"`
	Hostname  string         `json:"hostname"`
	Env       string         `json:"env"`
	Timestamp time.Time      `json:"timestamp"`
	Metadata  map[string]any `json:metadata`
}
