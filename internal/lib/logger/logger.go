package logger

import (
	"log/slog"
	"os"
	"time"

	"charm.land/log/v2"
)

func Load() {
	handler := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		TimeFormat:      time.DateTime, // e.g. "4:57PM"
		ReportCaller:    true,          // Set true if you want file:line numbers
		Level:           log.InfoLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
