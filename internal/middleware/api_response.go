package middleware

import (
	"encoding/json"
	"errors"
	"freshman-adaptation/internal/apperrors"
	"log"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("error encoding JSON response: %v", err)
	}
}

func WriteError(w http.ResponseWriter, err error) {
	appErr := apperrors.FromSentinel(err)
	if appErr == nil {
		appErr = apperrors.Internal("internal server error")
	}

	log.Printf("request error: %v", err)
	WriteJSON(w, appErr.Status, errorResponse{Error: appErr.Message})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				WriteError(w, apperrors.Internal("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		log.Printf("%s %s -> %d", r.Method, r.URL.RequestURI(), recorder.status)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func IsAppError(err error) bool {
	var appErr *apperrors.AppError
	return errors.As(err, &appErr)
}
