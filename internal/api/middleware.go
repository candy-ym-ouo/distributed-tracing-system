package api

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	written, err := r.ResponseWriter.Write(body)
	r.bytes += written
	return written, err
}

func RequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		log.Printf("method=%s path=%s status=%d bytes=%d duration=%s",
			r.Method, r.URL.Path, recorder.status, recorder.bytes, time.Since(started))
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic: %v\n%s", recovered, debug.Stack())
				writeError(w, http.StatusInternalServerError, "internal server error", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Auth-Token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isUpload := r.Method == http.MethodPost &&
				(r.URL.Path == "/api/v1/span" || r.URL.Path == "/api/v1/spans")
			if token != "" && isUpload && r.Header.Get("X-Auth-Token") != token {
				writeError(w, http.StatusUnauthorized, "invalid auth token", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
