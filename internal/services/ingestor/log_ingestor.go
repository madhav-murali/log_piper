// // internal/services/ingestor/log_ingestor.go
package ingestor

// type LogIngestor struct {
// 	natsClient *nats.Client
// 	storage    storage.LogRepository
// 	config     *config.Config
// }

// func New(cfg *config.Config) (*LogIngestor, error) {
// 	nc, err := nats.Connect(cfg.NATS.URL)
// 	if err != nil {
// 		return nil, fmt.Errorf("nats connection failed: %w", err)
// 	}

// 	js, err := nc.JetStream()
// 	if err != nil {
// 		return nil, fmt.Errorf("jetstream failed: %w", err)
// 	}

// 	return &LogIngestor{
// 		natsClient: nc,
// 		storage:    storage.NewPostgresStorage(cfg.DB),
// 		config:     cfg,
// 	}, nil
// }

// func (li *LogIngestor) Run() error {
// 	// Subscribe to log streams
// 	_, err := li.natsClient.QueueSubscribe(
// 		"logs.>",
// 		"LOG_INGESTORS",
// 		li.handleLogMessage,
// 	)
// 	return err
// }

// func (li *LogIngestor) handleLogMessage(msg *nats.Msg) {
// 	var logEntry domain.LogEntry
// 	if err := json.Unmarshal(msg.Data, &logEntry); err != nil {
// 		log.Printf("Failed to unmarshal log: %v", err)
// 		return
// 	}

// 	// Basic processing
// 	logEntry.Timestamp = time.Now()
// 	logEntry.Processed = true

// 	if err := li.storage.Store(logEntry); err != nil {
// 		log.Printf("Failed to store log: %v", err)
// 		return
// 	}

// 	msg.Ack()
// }
