package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	Page    int
	PerPage int
}

// parsePagination parses pagination parameters from query string.
// Only supports page-based pagination (page, per_page).
func parsePagination(r *http.Request) (*PaginationParams, error) {
	// Parse page parameter
	page, err := queryParamInt(r, "page", 1)
	if err != nil {
		return nil, fmt.Errorf("invalid page parameter: %w", err)
	}
	if page < 1 {
		page = 1
	}

	// Parse per_page parameter
	perPage, err := queryParamInt(r, "per_page", 20)
	if err != nil {
		return nil, fmt.Errorf("invalid per_page parameter: %w", err)
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	return &PaginationParams{
		Page:    page,
		PerPage: perPage,
	}, nil
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

// writePagePaginatedResponse writes a page-based paginated response.
func writePagePaginatedResponse(w http.ResponseWriter, r *http.Request, data interface{}, total int64, page, perPage int) error {
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	if totalPages == 0 {
		totalPages = 1
	}

	// Add pagination headers (GitHub style)
	w.Header().Set("Link", buildLinkHeader(r, page, perPage, totalPages))

	// Return just the data - pagination info is in headers
	return WriteJSON(w, http.StatusOK, data)
}

// buildLinkHeader builds Link header for pagination (GitHub style).
func buildLinkHeader(r *http.Request, page, perPage, totalPages int) string {
	var links []string

	// Build base URL from request
	baseURL := fmt.Sprintf("%s://%s%s",
		getScheme(r),
		r.Host,
		strings.Split(r.URL.Path, "?")[0]) // Remove query parameters from path

	// GitHub style: only next and first links
	if page < totalPages {
		nextPage := page + 1
		links = append(links, fmt.Sprintf(`<%s?page=%d&per_page=%d>; rel="next"`, baseURL, nextPage, perPage))
	}

	// Always include first link (GitHub style)
	links = append(links, fmt.Sprintf(`<%s?page=1&per_page=%d>; rel="first"`, baseURL, perPage))

	return strings.Join(links, ", ")
}

// getScheme returns the scheme (http/https) from the request.
func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	if scheme := r.Header.Get("X-Forwarded-Scheme"); scheme != "" {
		return scheme
	}
	return "http"
}
