package middlewares

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
)

// gzipWriter gzip writer.
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write writes using custom writer.
func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GzipMiddleware shrinks response with gzip.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		defer func() {
			if err = gz.Close(); err != nil {
				log.Print(err)
			}
		}()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
