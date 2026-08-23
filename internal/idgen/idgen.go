package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func TraceID() (string, error) {
	return randomHex(16)
}

func SpanID() (string, error) {
	return randomHex(8)
}

func MustTraceID() string {
	id, err := TraceID()
	if err != nil {
		panic(fmt.Sprintf("generate trace id: %v", err))
	}
	return id
}

func MustSpanID() string {
	id, err := SpanID()
	if err != nil {
		panic(fmt.Sprintf("generate span id: %v", err))
	}
	return id
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
