package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// WriteJSON writes a JSON response.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// WriteJSONIgnoreError writes a JSON response and ignores any encoding errors.
func WriteJSONIgnoreError(w http.ResponseWriter, status int, data interface{}) error {
	return WriteJSON(w, status, data)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, message string) error {
	errorResponse := map[string]string{
		"error": message,
	}
	return WriteJSON(w, status, errorResponse)
}

// writeValidationError writes a validation error response.
func writeValidationError(w http.ResponseWriter, message string) error {
	return writeError(w, http.StatusBadRequest, message)
}

// writeNotFound writes a not found error response.
func writeNotFound(w http.ResponseWriter, resource string) error {
	return writeError(w, http.StatusNotFound, fmt.Sprintf("%s not found", resource))
}

// writeInternalError writes an internal server error response.
func writeInternalError(w http.ResponseWriter, err error) error {
	return writeError(w, http.StatusInternalServerError, "Internal server error")
}

// param extracts a URL parameter as string.
func param(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

// paramUUID extracts a URL parameter as UUID.
func paramUUID(r *http.Request, name string) (uuid.UUID, error) {
	paramStr := chi.URLParam(r, name)
	if paramStr == "" {
		return uuid.Nil, fmt.Errorf("parameter %s is required", name)
	}
	return uuid.Parse(paramStr)
}

// queryParam extracts a query parameter as string.
func queryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// queryParamInt extracts a query parameter as int.
func queryParamInt(r *http.Request, name string, defaultValue int) (int, error) {
	paramStr := r.URL.Query().Get(name)
	if paramStr == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(paramStr)
}

// queryParamBool extracts a query parameter as bool.
func queryParamBool(r *http.Request, name string, defaultValue bool) (bool, error) {
	paramStr := r.URL.Query().Get(name)
	if paramStr == "" {
		return defaultValue, nil
	}
	return strconv.ParseBool(paramStr)
}

// parseJSON parses JSON from request body.
func parseJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is required")
	}
	defer func() { _ = r.Body.Close() }()
	return json.NewDecoder(r.Body).Decode(v)
}

// PaginationParams represents pagination parameters.
type PaginationParams struct {
	Limit  int
	Offset int
}

// parsePagination parses pagination parameters from query string.
func parsePagination(r *http.Request) (*PaginationParams, error) {
	limit, err := queryParamInt(r, "limit", 20)
	if err != nil {
		return nil, fmt.Errorf("invalid limit parameter: %w", err)
	}
	if limit < 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := queryParamInt(r, "offset", 0)
	if err != nil {
		return nil, fmt.Errorf("invalid offset parameter: %w", err)
	}
	if offset < 0 {
		offset = 0
	}

	return &PaginationParams{
		Limit:  limit,
		Offset: offset,
	}, nil
}

// PaginatedResponse represents a paginated response.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	HasMore    bool        `json:"has_more"`
	NextOffset *int        `json:"next_offset,omitempty"`
}

// writePaginatedResponse writes a paginated response.
func writePaginatedResponse(w http.ResponseWriter, data interface{}, total int64, limit, offset int) error {
	hasMore := int64(offset+limit) < total
	var nextOffset *int
	if hasMore {
		next := offset + limit
		nextOffset = &next
	}

	response := PaginatedResponse{
		Data:       data,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}

	return WriteJSON(w, http.StatusOK, response)
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services,omitempty"`
}

// writeHealthResponse writes a health check response.
func writeHealthResponse(w http.ResponseWriter, status string, services map[string]string) error {
	response := HealthResponse{
		Status:    status,
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
		Version:   "1.0.0",
		Services:  services,
	}
	return WriteJSON(w, http.StatusOK, response)
}
