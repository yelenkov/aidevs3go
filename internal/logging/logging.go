package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Setup() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal().Err(err).Msg("Can't create logs directory")
	}

	// Open a file for writing logs
	logFile, err := os.OpenFile(
		"logs/app.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0664,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not open log file")
	}

	// Configure zerolog
	multi := zerolog.MultiLevelWriter(os.Stdout, logFile)

	// Basic JSON configuration
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Logger = zerolog.New(multi).
		With().
		Timestamp().
		Str("app", "cenzura").
		Logger()
}
