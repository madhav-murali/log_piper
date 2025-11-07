package storage

import (
	"context"

	"github.com/bluebox/internal/domain"

	"github.com/jackc/pgx/v4/pgxpool"
)

type TimescaleStorage struct {
	pool *pgxpool.Pool
}

func NewTimescaleStorage(connString string) (*TimescaleStorage, error) {
	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	return &TimescaleStorage{pool: pool}, nil
}

func (s *TimescaleStorage) InsertMetric(ctx context.Context, metric *domain.MetricEntry) error {
	query := `INSERT INTO metrics (id, timestamp, service_name, metric_name, value, metadata) 
	          VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.pool.Exec(ctx, query, metric.Id, metric.Timestamp, metric.Service, metric.Metric, metric.Value, metric.Env)
	return err
}

func (s *TimescaleStorage) Init(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS metrics (
		id TEXT PRIMARY KEY,
		timestamp TIMESTAMP WITH TIME ZONE,
		service_name TEXT,
		metric_name TEXT,
		value DOUBLE PRECISION,
		env TEXT
	)`

	_, err := s.pool.Exec(ctx, query)
	if err != nil {
		return err
	}

	// Convert to hypertable if it's not already
	query = "SELECT create_hypertable('metrics', 'timestamp', if_not_exists => TRUE)"
	_, err = s.pool.Exec(ctx, query)
	return err
}

func (s *TimescaleStorage) Close() {
	s.pool.Close()
}
