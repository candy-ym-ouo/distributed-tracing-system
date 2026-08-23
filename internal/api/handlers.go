package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"distributed-tracing-system/internal/aggregation"
	"distributed-tracing-system/internal/collector"
	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/query"
	"distributed-tracing-system/internal/storage"
)

type handlers struct {
	collector *collector.Collector
	query     *query.Service
	stats     *aggregation.Aggregator
	storage   storage.Storage
}

type envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (h *handlers) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{Message: "ok", Data: map[string]any{
		"status": "up", "storage": h.storage.Stats(),
	}})
}

func (h *handlers) singleSpan(w http.ResponseWriter, r *http.Request) {
	var span model.Span
	if err := decodeBody(r, &span); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.collector.Submit([]model.Span{span}); err != nil {
		handleSubmitError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, envelope{Message: "ok", Data: map[string]int{"accepted": 1}})
}

func (h *handlers) batchSpans(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Spans []model.Span `json:"spans"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.collector.Submit(body.Spans); err != nil {
		handleSubmitError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, envelope{Message: "ok", Data: map[string]int{"accepted": len(body.Spans)}})
}

func (h *handlers) trace(w http.ResponseWriter, r *http.Request) {
	traceID := r.PathValue("traceID")
	if len(traceID) != 32 {
		writeError(w, http.StatusBadRequest, "invalid trace id", nil)
		return
	}
	if _, err := hex.DecodeString(traceID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid trace id", nil)
		return
	}
	tree, err := h.query.Trace(r.Context(), traceID)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error(), nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, envelope{Message: "ok", Data: tree})
}

func (h *handlers) traces(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	queryValue := model.Query{
		StartTime: int64Value(values.Get("start")),
		EndTime:   int64Value(values.Get("end")),
		Service:   values.Get("service"),
		Operation: values.Get("operation"),
		Status:    model.Status(values.Get("status")),
		Page:      int(int64Value(values.Get("page"))),
		PageSize:  int(int64Value(values.Get("page"))),
	}
	page, err := h.query.Search(r.Context(), queryValue)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, envelope{Message: "ok", Data: page})
}

func (h *handlers) serviceStats(w http.ResponseWriter, r *http.Request) {
	start, end := rangeValues(r)
	result, err := h.stats.Services(r.Context(), start, end)
	writeResult(w, result, err)
}

func (h *handlers) operationStats(w http.ResponseWriter, r *http.Request) {
	start, end := rangeValues(r)
	limit := int(int64Value(r.URL.Query().Get("limit")))
	result, err := h.stats.Operations(r.Context(), start, end, limit)
	writeResult(w, result, err)
}

func (h *handlers) topology(w http.ResponseWriter, r *http.Request) {
	start, end := rangeValues(r)
	result, err := h.stats.Topology(r.Context(), start, end)
	writeResult(w, result, err)
}

func (h *handlers) overview(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{Message: "ok", Data: map[string]any{
		"collector": h.collector.Metrics(),
		"storage":   h.storage.Stats(),
	}})
}

func decodeBody(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 8<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func handleSubmitError(w http.ResponseWriter, err error) {
	var invalid *collector.InvalidBatchError
	switch {
	case errors.As(err, &invalid):
		writeError(w, http.StatusBadRequest, invalid.Error(), invalid.Details)
	case errors.Is(err, collector.ErrQueueFull):
		writeError(w, http.StatusTooManyRequests, err.Error(), nil)
	default:
		writeError(w, http.StatusServiceUnavailable, err.Error(), nil)
	}
}

func writeResult(w http.ResponseWriter, result any, err error) {
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, envelope{Message: "ok", Data: result})
}

func writeError(w http.ResponseWriter, status int, message string, details any) {
	writeJSON(w, status, envelope{Code: status, Message: message, Data: details})
}

func writeJSON(w http.ResponseWriter, status int, body envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func rangeValues(r *http.Request) (int64, int64) {
	return int64Value(r.URL.Query().Get("start")), int64Value(r.URL.Query().Get("end"))
}

func int64Value(value string) int64 {
	if value == "" {
		return 0
	}
	parsed, _ := strconv.ParseInt(value, 10, 64)
	return parsed
}

func unixNano(t time.Time) int64 { return t.UnixNano() }
