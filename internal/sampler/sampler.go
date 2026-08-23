package sampler

import (
	"errors"
	"strings"
)

type Sampler interface {
	Sample(traceID string) bool
}

type Always struct{}

func (Always) Sample(string) bool { return true }

type Never struct{}

func (Never) Sample(string) bool { return false }

func New(kind string, value float64) (Sampler, error) {
	switch strings.ToLower(kind) {
	case "always", "":
		return Always{}, nil
	case "never":
		return Never{}, nil
	case "rate":
		return NewRate(value)
	case "limit":
		return NewRateLimit(int(value)), nil
	default:
		return nil, errors.New("unknown sampling strategy")
	}
}
