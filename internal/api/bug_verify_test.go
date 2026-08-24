package api

import (
	"distributed-tracing-system/internal/aggregation"
	"distributed-tracing-system/internal/collector"
	"distributed-tracing-system/internal/query"
	"distributed-tracing-system/internal/storage"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBug10PageSizePassesThrough(t *testing.T) {
	s := storage.NewMemory()
	c := collector.New(s, 1, 2)
	defer c.Close(t.Context())
	h := NewRouter(Dependencies{Collector: c, Query: query.New(s), Stats: aggregation.New(s), Storage: s})
	r := httptest.NewRequest(http.MethodGet, "/api/v1/traces?page=2&pageSize=7", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var body struct {
		Data struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Page != 2 || body.Data.PageSize != 7 {
		t.Fatalf("body=%s", w.Body.String())
	}
}
