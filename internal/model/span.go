package model

import (
	"errors"
	"strings"
	"time"
)

type SpanKind string

const (
	KindServer   SpanKind = "server"
	KindClient   SpanKind = "client"
	KindProducer SpanKind = "producer"
	KindConsumer SpanKind = "consumer"
)

type Status string

const (
	StatusOK    Status = "ok"
	StatusError Status = "error"
)

type Span struct {
	TraceID       string            `json:"trace_id"`
	SpanID        string            `json:"span_id"`
	ParentSpanID  string            `json:"parent_span_id,omitempty"`
	ServiceName   string            `json:"service_name"`
	OperationName string            `json:"operation_name"`
	Kind          SpanKind          `json:"kind"`
	StartTime     int64             `json:"start_time"`
	EndTime       int64             `json:"end_time"`
	DurationUs    int64             `json:"duration_us"`
	Tags          map[string]string `json:"tags,omitempty"`
	Status        Status            `json:"status"`
	Sampled       bool              `json:"sampled"`
	Forced        bool              `json:"forced,omitempty"`
}

func (s *Span) Finish(at time.Time, err error) {
	s.EndTime = at.UnixNano()
	if s.EndTime < s.StartTime {
		s.EndTime = s.StartTime
	}
	s.DurationUs = (s.EndTime - s.StartTime) / int64(time.Microsecond)
	if err == nil {
		s.Status = StatusOK
		return
	}
	s.Status = StatusError
	s.Forced = true
	if s.Tags == nil {
		s.Tags = make(map[string]string)
	}
	s.Tags["error.message"] = err.Error()
}

func (s *Span) SetTag(key, value string) {
	if s.Tags == nil {
		s.Tags = make(map[string]string)
	}
	s.Tags[key] = value
}

func (s Span) Validate() error {
	if len(s.TraceID) != 32 || !isHex(s.TraceID) {
		return errors.New("trace_id must be 32 hexadecimal characters")
	}
	if len(s.SpanID) != 16 || !isHex(s.SpanID) {
		return errors.New("span_id must be 16 hexadecimal characters")
	}
	if s.ParentSpanID != "" && (len(s.ParentSpanID) != 16 || !isHex(s.ParentSpanID)) {
		return errors.New("parent_span_id must be empty or 16 hexadecimal characters")
	}
	if strings.TrimSpace(s.ServiceName) == "" {
		return errors.New("service_name is required")
	}
	if strings.TrimSpace(s.OperationName) == "" {
		return errors.New("operation_name is required")
	}
	if s.StartTime <= 0 {
		return errors.New("start_time must be positive")
	}
	if s.EndTime > 0 && s.EndTime < s.StartTime {
		return errors.New("end_time cannot precede start_time")
	}
	return nil
}

func isHex(value string) bool {
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func CloneSpan(in Span) Span {
	out := in
	if in.Tags != nil {
		out.Tags = make(map[string]string, len(in.Tags))
		for key, value := range in.Tags {
			out.Tags[key] = value
		}
	}
	return out
}
