package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"distributed-tracing-system/internal/aggregation"
	"distributed-tracing-system/internal/collector"
	"distributed-tracing-system/internal/query"
	"distributed-tracing-system/internal/storage"
)

type Dependencies struct {
	Collector *collector.Collector
	Query     *query.Service
	Stats     *aggregation.Aggregator
	Storage   storage.Storage
	WebDir    string
	AuthToken string
}

func NewRouter(deps Dependencies) http.Handler {
	h := &handlers{
		collector: deps.Collector,
		query:     deps.Query,
		stats:     deps.Stats,
		storage:   deps.Storage,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /api/v1/span", h.singleSpan)
	mux.HandleFunc("POST /api/v1/spans", h.batchSpans)
	mux.HandleFunc("GET /api/v1/traces/{traceID}", h.trace)
	mux.HandleFunc("GET /api/v1/traces", h.traces)
	mux.HandleFunc("GET /api/v1/stats/services", h.serviceStats)
	mux.HandleFunc("GET /api/v1/stats/operations", h.operationStats)
	mux.HandleFunc("GET /api/v1/stats/topology", h.topology)
	mux.HandleFunc("GET /api/v1/stats/overview", h.overview)
	mux.Handle("GET /", webHandler(deps.WebDir))
	var handler http.Handler = mux
	handler = Auth(deps.AuthToken)(handler)
	handler = CORS(handler)
	handler = Recover(handler)
	handler = RequestLog(handler)
	return handler
}

func webHandler(directory string) http.Handler {
	if directory == "" {
		return http.NotFoundHandler()
	}
	absolute, err := filepath.Abs(directory)
	if err != nil {
		return http.NotFoundHandler()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimPrefix(filepath.Clean(r.URL.Path), string(filepath.Separator))
		if clean == "." || clean == "" {
			clean = "index.html"
		}
		target := filepath.Join(absolute, clean)
		if !strings.HasPrefix(target, absolute+string(filepath.Separator)) {
			http.NotFound(w, r)
			return
		}
		info, err := os.Stat(target)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, target)
	})
}
