package collector

import (
	"fmt"

	"distributed-tracing-system/internal/model"
)

const MaxBatchSize = 512

type ValidationError struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

type InvalidBatchError struct {
	Details []ValidationError `json:"details"`
}

func (e *InvalidBatchError) Error() string {
	return fmt.Sprintf("batch contains %d invalid spans", len(e.Details))
}

func ValidateBatch(spans []model.Span) error {
	if len(spans) == 0 {
		return &InvalidBatchError{Details: []ValidationError{{Index: -1, Error: "batch is empty"}}}
	}
	if len(spans) > MaxBatchSize {
		return &InvalidBatchError{Details: []ValidationError{{Index: -1, Error: "batch exceeds 512 spans"}}}
	}
	details := make([]ValidationError, 0)
	for index := range spans {
		if err := spans[index].Validate(); err != nil {
			details = append(details, ValidationError{Index: index, Error: err.Error()})
		}
	}
	if len(details) > 0 {
		return &InvalidBatchError{Details: details}
	}
	return nil
}

func Deduplicate(spans []model.Span) []model.Span {
	positions := make(map[string]int, len(spans))
	result := make([]model.Span, 0, len(spans))
	for _, span := range spans {
		key := span.TraceID + ":" + span.SpanID
		if index, exists := positions[key]; exists {
			result[index] = span
			continue
		}
		positions[key] = len(result)
		result = append(result, span)
	}
	return result
}

func GroupByTrace(spans []model.Span) map[string][]model.Span {
	groups := make(map[string][]model.Span)
	for _, span := range spans {
		groups[span.TraceID] = append(groups[span.TraceID], span)
	}
	return groups
}
