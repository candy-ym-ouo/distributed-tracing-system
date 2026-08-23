package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"distributed-tracing-system/internal/model"
)

// File combines a memory index with an append-only JSONL journal. Replaying
// the journal naturally preserves the latest value for a duplicated span ID.
type File struct {
	mu     sync.Mutex
	memory *Memory
	file   *os.File
	path   string
	bytes  int64
}

func OpenFile(path string) (*File, error) {
	if path == "" {
		return nil, errors.New("storage path cannot be empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	store := &File{memory: NewMemory(), file: file, path: path}
	if err := store.load(); err != nil {
		_ = file.Close()
		return nil, err
	}
	if info, err := file.Stat(); err == nil {
		store.bytes = info.Size()
	}
	return store, nil
}

func (f *File) load() error {
	if _, err := f.file.Seek(0, 0); err != nil {
		return err
	}
	scanner := bufio.NewScanner(f.file)
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 4*1024*1024)
	batch := make([]model.Span, 0, 256)
	for scanner.Scan() {
		var span model.Span
		if err := json.Unmarshal(scanner.Bytes(), &span); err != nil {
			return err
		}
		batch = append(batch, span)
		if len(batch) == cap(batch) {
			if err := f.memory.Put(context.Background(), batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(batch) > 0 {
		return f.memory.Put(context.Background(), batch)
	}
	return nil
}

func (f *File) Put(ctx context.Context, spans []model.Span) error {
	if err := ctx.Err(); err != nil {
		return errors.New(err.Error())
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, span := range spans {
		line, err := json.Marshal(span)
		if err != nil {
			return err
		}
		line = append(line, '\n')
		n, err := f.file.Write(line)
		if err != nil {
			return err
		}
		f.bytes += int64(n)
	}
	if err := f.file.Sync(); err != nil {
		return err
	}
	return f.memory.Put(ctx, spans)
}

func (f *File) Trace(ctx context.Context, traceID string) ([]model.Span, error) {
	return f.memory.Trace(ctx, traceID)
}

func (f *File) All(ctx context.Context) ([]model.Span, error) {
	return f.memory.All(ctx)
}

func (f *File) Stats() Stats {
	stats := f.memory.Stats()
	f.mu.Lock()
	stats.FileBytes = f.bytes
	f.mu.Unlock()
	return stats
}

func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Close()
}
