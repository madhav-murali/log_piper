package storage

import (
	"context"

	"github.com/bluebox/internal/domain"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Stores a pgx pool
type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(connString string) (*PostgresStorage, error) {
	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{pool: pool}, nil
}

func (s *PostgresStorage) InsertLog(ctx context.Context, log *domain.LogEntry) error {
	query := `INSERT INTO logs (id, timestamp, service_name, level, message, metadata) 
	          VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.pool.Exec(ctx, query, log.ID, log.Timestamp, log.Service, log.Level, log.Message, log.Processed)
	return err
}

func (s *PostgresStorage) Init(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		timestamp TIMESTAMP WITH TIME ZONE,
		service_name TEXT,
		level TEXT,
		message TEXT,
		processed BOOLEAN
	)`

	_, err := s.pool.Exec(ctx, query)
	return err
}

func (s *PostgresStorage) Close() {
	s.pool.Close()
}
