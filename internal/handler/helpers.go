package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/SmoothWay/metrics/internal/logger"
)

type envelope map[string]any

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")

	js, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}
	js = append(js, '\n')

	w.WriteHeader(status)
	w.Write(js)
}

func errorResponse(w http.ResponseWriter, r *http.Request, status int, err error, message any) {
	env := envelope{"error": message}
	logger.Log.Error("error handling request", zap.Int("status", status), zap.String("url", r.URL.String()), zap.Error(err))

	writeJSON(w, status, env)
}
func badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	message := "bad request"
	errorResponse(w, r, http.StatusBadRequest, err, message)
}

func serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	message := "the server encountered a problem and could not process your request"
	errorResponse(w, r, http.StatusInternalServerError, err, message)
}

func notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the required resource could not be found"
	env := envelope{"error": message}

	writeJSON(w, http.StatusNotFound, env)
}

func methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	env := envelope{"error": message}

	writeJSON(w, http.StatusMethodNotAllowed, env)
}
