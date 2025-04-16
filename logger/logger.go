package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type APIHandler struct {
	Endpoint string
	Level    slog.Level
	attrs    []slog.Attr
}

func (h *APIHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.Level
}

func (h *APIHandler) Handle(ctx context.Context, record slog.Record) error {
	data := map[string]interface{}{
		"time":    record.Time.Format(time.RFC3339),
		"level":   record.Level.String(),
		"message": record.Message,
	}

	for _, item := range h.attrs {
		data[item.Key] = item.Value.Any()
	}

	record.Attrs(func(a slog.Attr) bool {
		data[a.Key] = a.Value.Any()
		return true
	})

	fmt.Println(data)
	// Convert to JSON
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Send to API
	req, err := http.NewRequest("POST", h.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (h *APIHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *APIHandler) WithGroup(name string) slog.Handler {
	return h
}

func Init(defaultAttrs []slog.Attr) *slog.Logger {
	apiEndpoint := os.Getenv("REMOTE_TRIGGER_LOGGERS_ENDPOINT")
	handler := &APIHandler{
		Endpoint: apiEndpoint,
		Level:    slog.LevelInfo,
		attrs:    defaultAttrs,
	}

	logger := slog.New(handler)

	return logger
}
