package handlers

import (
	"net/http"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// QuoteHandler handles HTTP requests for quotes.
type QuoteHandler struct {
	service *services.QuoteService
}

// NewQuoteHandler creates a new QuoteHandler.
func NewQuoteHandler(service *services.QuoteService) *QuoteHandler {
	return &QuoteHandler{
		service: service,
	}
}

// ListQuotes handles GET /v1/quotes.
func (h *QuoteHandler) ListQuotes(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	pagination, err := parsePagination(r)
	if err != nil {
		_ = writeValidationError(w, err.Error())
		return
	}

	// Create quote list options
	opts := models.NewQuoteListOptions()
	opts.Page = pagination.Page
	opts.PerPage = pagination.PerPage

	// Parse filter parameters
	if userIDStr := queryParam(r, "user_id"); userIDStr != "" {
		if id, err := uuid.Parse(userIDStr); err == nil {
			opts.UserID = &id
		}
	}

	if productIDStr := queryParam(r, "product_id"); productIDStr != "" {
		if id, err := uuid.Parse(productIDStr); err == nil {
			opts.ProductID = &id
		}
	}

	opts.Status = queryParam(r, "status")
	opts.Currency = queryParam(r, "currency")

	// Get quotes
	quotes, err := h.service.ListQuotes(r.Context(), opts)
	if err != nil {
		_ = writeInternalError(w, err)
		return
	}

	// Get total count
	total, err := h.service.CountQuotes(r.Context(), opts)
	if err != nil {
		_ = writeInternalError(w, err)
		return
	}

	// Write page-based paginated response
	if err := writePagePaginatedResponse(w, r, quotes, total, opts.Page, opts.PerPage); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// GetQuote handles GET /v1/quotes/{id}.
func (h *QuoteHandler) GetQuote(w http.ResponseWriter, r *http.Request) {
	// Parse quote ID
	quoteID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid quote ID")
		return
	}

	// Get quote
	quote, err := h.service.GetQuote(r.Context(), quoteID)
	if err != nil {
		if err.Error() == "quote not found" {
			_ = writeNotFound(w, "Quote")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	if err := WriteJSONIgnoreError(w, http.StatusOK, quote); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// GetQuoteByNumber handles GET /v1/quotes/number/{number}.
func (h *QuoteHandler) GetQuoteByNumber(w http.ResponseWriter, r *http.Request) {
	// Parse quote number
	quoteNumber := param(r, "number")
	if quoteNumber == "" {
		_ = writeValidationError(w, "Quote number is required")
		return
	}

	// Get quote
	quote, err := h.service.GetQuoteByNumber(r.Context(), quoteNumber)
	if err != nil {
		if err.Error() == "quote not found" {
			_ = writeNotFound(w, "Quote")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	if err := WriteJSONIgnoreError(w, http.StatusOK, quote); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// CreateQuote handles POST /v1/quotes.
func (h *QuoteHandler) CreateQuote(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var quote models.Quote
	if err := parseJSON(r, &quote); err != nil {
		_ = writeValidationError(w, "Invalid JSON body")
		return
	}

	// Create quote
	if err := h.service.CreateQuote(r.Context(), &quote); err != nil {
		_ = writeValidationError(w, err.Error())
		return
	}

	// Write response
	if err := WriteJSONIgnoreError(w, http.StatusCreated, quote); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// UpdateQuote handles PUT /v1/quotes/{id}.
func (h *QuoteHandler) UpdateQuote(w http.ResponseWriter, r *http.Request) {
	// Parse quote ID
	quoteID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid quote ID")
		return
	}

	// Parse request body
	var quote models.Quote
	if err := parseJSON(r, &quote); err != nil {
		_ = writeValidationError(w, "Invalid JSON body")
		return
	}

	// Set ID from URL
	quote.ID = quoteID

	// Update quote
	if err := h.service.UpdateQuote(r.Context(), &quote); err != nil {
		if err.Error() == "quote not found" {
			_ = writeNotFound(w, "Quote")
			return
		}
		_ = writeValidationError(w, err.Error())
		return
	}

	// Write response
	if err := WriteJSONIgnoreError(w, http.StatusOK, quote); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// DeleteQuote handles DELETE /v1/quotes/{id}.
func (h *QuoteHandler) DeleteQuote(w http.ResponseWriter, r *http.Request) {
	// Parse quote ID
	quoteID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid quote ID")
		return
	}

	// Delete quote
	if err := h.service.DeleteQuote(r.Context(), quoteID); err != nil {
		if err.Error() == "quote not found" {
			_ = writeNotFound(w, "Quote")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	w.WriteHeader(http.StatusNoContent)
}

// ExpireQuote handles POST /v1/quotes/{id}/expire.
func (h *QuoteHandler) ExpireQuote(w http.ResponseWriter, r *http.Request) {
	// Parse quote ID
	quoteID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid quote ID")
		return
	}

	// Expire quote
	if err := h.service.ExpireQuote(r.Context(), quoteID); err != nil {
		if err.Error() == "quote not found" {
			_ = writeNotFound(w, "Quote")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes registers quote routes with the router.
func (h *QuoteHandler) RegisterRoutes(r chi.Router) {
	r.Route("/quotes", func(r chi.Router) {
		r.Get("/", h.ListQuotes)
		r.Post("/", h.CreateQuote)
		r.Get("/{id}", h.GetQuote)
		r.Put("/{id}", h.UpdateQuote)
		r.Delete("/{id}", h.DeleteQuote)
		r.Post("/{id}/expire", h.ExpireQuote)
		r.Get("/number/{number}", h.GetQuoteByNumber)
	})
}
