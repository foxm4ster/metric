package metric

import "net/http"

type statusResponseWriter struct {
	http.ResponseWriter
	code int
}

func (w statusResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}
