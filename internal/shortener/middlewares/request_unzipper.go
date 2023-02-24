package middlewares

import (
	"compress/gzip"
	"log"
	"net/http"
)

// RequestUnzip handles gzipped request.
func RequestUnzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			defer func() {
				if err = reader.Close(); err != nil {
					log.Print(err)
				}
				if err = r.Body.Close(); err != nil {
					log.Print(err)
				}
			}()
			r.Body = reader
		}

		next.ServeHTTP(w, r)
	})
}
