package recorder

import "net/http"

type StatusRecorder struct {
	http.ResponseWriter
	StatusCode int
}

func RecordStatus(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{w, http.StatusOK}
}

func (rs *StatusRecorder) WriteHeader(code int) {
	rs.StatusCode = code
	rs.ResponseWriter.WriteHeader(code)
}
