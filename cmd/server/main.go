package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"distributed-tracing-system/internal/aggregation"
	"distributed-tracing-system/internal/api"
	"distributed-tracing-system/internal/collector"
	"distributed-tracing-system/internal/config"
	"distributed-tracing-system/internal/idgen"
	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/query"
	"distributed-tracing-system/internal/storage"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	cfg, err := config.Load(args)
	if err != nil {
		return err
	}
	store, err := storage.OpenFile(cfg.DataFile)
	if err != nil {
		return err
	}
	collect := collector.New(store, cfg.Workers, cfg.QueueSize)
	queryService := query.New(store)
	aggregator := aggregation.New(store)
	router := api.NewRouter(api.Dependencies{
		Collector: collect,
		Query:     queryService,
		Stats:     aggregator,
		Storage:   store,
		WebDir:    cfg.WebDir,
		AuthToken: cfg.AuthToken,
	})
	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if cfg.Demo {
		go generateDemo(ctx, collect)
	}
	serveError := make(chan error, 1)
	go func() {
		log.Printf("distributed tracing server listening on %s", cfg.Addr)
		serveError <- server.ListenAndServe()
	}()
	select {
	case err := <-serveError:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdown); err != nil {
		return err
	}
	return collect.Close(shutdown)
}

func generateDemo(ctx context.Context, collect *collector.Collector) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			_ = collect.Submit(demoTrace(now))
		}
	}
}

func demoTrace(now time.Time) []model.Span {
	traceID := idgen.MustTraceID()
	rootID := idgen.MustSpanID()
	clientID := idgen.MustSpanID()
	rootDuration := int64(20_000 + rand.Intn(80_000))
	paymentDuration := int64(5_000 + rand.Intn(30_000))
	status := model.StatusOK
	if rand.Intn(10) == 0 {
		status = model.StatusError
	}
	root := demoSpan(traceID, rootID, "", "order", "POST /orders", model.KindServer,
		now.UnixNano(), rootDuration, status)
	client := demoSpan(traceID, clientID, rootID, "order", "POST payment", model.KindClient,
		now.Add(2*time.Millisecond).UnixNano(), paymentDuration+2_000, status)
	payment := demoSpan(traceID, idgen.MustSpanID(), clientID, "payment", "POST /charge", model.KindServer,
		now.Add(4*time.Millisecond).UnixNano(), paymentDuration, status)
	inventory := demoSpan(traceID, idgen.MustSpanID(), rootID, "inventory", "GET /stock", model.KindServer,
		now.Add(3*time.Millisecond).UnixNano(), int64(3_000+rand.Intn(12_000)), model.StatusOK)
	return []model.Span{payment, root, inventory, client}
}

func demoSpan(traceID, spanID, parentID, service, operation string, kind model.SpanKind,
	start, duration int64, status model.Status) model.Span {
	return model.Span{
		TraceID: traceID, SpanID: spanID, ParentSpanID: parentID,
		ServiceName: service, OperationName: operation, Kind: kind,
		StartTime: start, EndTime: start + duration*int64(time.Microsecond),
		DurationUs: duration, Status: status, Sampled: true,
		Tags: map[string]string{"demo": "true"},
	}
}
