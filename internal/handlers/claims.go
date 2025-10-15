package handlers

import (
	"net/http"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ClaimHandler handles HTTP requests for claims.
type ClaimHandler struct {
	service *services.ClaimService
}

// NewClaimHandler creates a new ClaimHandler.
func NewClaimHandler(service *services.ClaimService) *ClaimHandler {
	return &ClaimHandler{
		service: service,
	}
}

// ListClaims handles GET /v1/claims.
func (h *ClaimHandler) ListClaims(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	pagination, err := parsePagination(r)
	if err != nil {
		_ = writeValidationError(w, err.Error())
		return
	}

	// Parse filter parameters
	var userID *uuid.UUID
	if userIDStr := queryParam(r, "user_id"); userIDStr != "" {
		if id, err := uuid.Parse(userIDStr); err == nil {
			userID = &id
		}
	}

	var policyID *uuid.UUID
	if policyIDStr := queryParam(r, "policy_id"); policyIDStr != "" {
		if id, err := uuid.Parse(policyIDStr); err == nil {
			policyID = &id
		}
	}

	status := queryParam(r, "status")

	// Get claims
	claims, err := h.service.ListClaims(r.Context(), userID, policyID, status, pagination.Limit, pagination.Offset)
	if err != nil {
		_ = writeInternalError(w, err)
		return
	}

	// Get total count
	total, err := h.service.CountClaims(r.Context(), userID, policyID, status)
	if err != nil {
		_ = writeInternalError(w, err)
		return
	}

	// Write page-based paginated response
	if err := writePagePaginatedResponse(w, r, claims, total, pagination.Page, pagination.PerPage); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// GetClaim handles GET /v1/claims/{id}.
func (h *ClaimHandler) GetClaim(w http.ResponseWriter, r *http.Request) {
	// Parse claim ID
	claimID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid claim ID")
		return
	}

	// Get claim
	claim, err := h.service.GetClaim(r.Context(), claimID)
	if err != nil {
		if err.Error() == "claim not found" {
			_ = writeNotFound(w, "Claim")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	if err := WriteJSONIgnoreError(w, http.StatusOK, claim); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// GetClaimByNumber handles GET /v1/claims/number/{number}.
func (h *ClaimHandler) GetClaimByNumber(w http.ResponseWriter, r *http.Request) {
	// Parse claim number
	claimNumber := param(r, "number")
	if claimNumber == "" {
		_ = writeValidationError(w, "Claim number is required")
		return
	}

	// Get claim
	claim, err := h.service.GetClaimByNumber(r.Context(), claimNumber)
	if err != nil {
		if err.Error() == "claim not found" {
			_ = writeNotFound(w, "Claim")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	if err := WriteJSONIgnoreError(w, http.StatusOK, claim); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// CreateClaim handles POST /v1/claims.
func (h *ClaimHandler) CreateClaim(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var claim models.Claim
	if err := parseJSON(r, &claim); err != nil {
		_ = writeValidationError(w, "Invalid JSON body")
		return
	}

	// Create claim
	if err := h.service.CreateClaim(r.Context(), &claim); err != nil {
		_ = writeValidationError(w, err.Error())
		return
	}

	// Write response
	if err := WriteJSONIgnoreError(w, http.StatusCreated, claim); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// UpdateClaim handles PUT /v1/claims/{id}.
func (h *ClaimHandler) UpdateClaim(w http.ResponseWriter, r *http.Request) {
	// Parse claim ID
	claimID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid claim ID")
		return
	}

	// Parse request body
	var claim models.Claim
	if err := parseJSON(r, &claim); err != nil {
		_ = writeValidationError(w, "Invalid JSON body")
		return
	}

	// Set ID from URL
	claim.ID = claimID

	// Update claim
	if err := h.service.UpdateClaim(r.Context(), &claim); err != nil {
		if err.Error() == "claim not found" {
			_ = writeNotFound(w, "Claim")
			return
		}
		_ = writeValidationError(w, err.Error())
		return
	}

	// Write response
	if err := WriteJSONIgnoreError(w, http.StatusOK, claim); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// DeleteClaim handles DELETE /v1/claims/{id}.
func (h *ClaimHandler) DeleteClaim(w http.ResponseWriter, r *http.Request) {
	// Parse claim ID
	claimID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid claim ID")
		return
	}

	// Delete claim
	if err := h.service.DeleteClaim(r.Context(), claimID); err != nil {
		if err.Error() == "claim not found" {
			_ = writeNotFound(w, "Claim")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes registers claim routes with the router.
func (h *ClaimHandler) RegisterRoutes(r chi.Router) {
	r.Route("/claims", func(r chi.Router) {
		r.Get("/", h.ListClaims)
		r.Post("/", h.CreateClaim)
		r.Get("/{id}", h.GetClaim)
		r.Put("/{id}", h.UpdateClaim)
		r.Delete("/{id}", h.DeleteClaim)
		r.Get("/number/{number}", h.GetClaimByNumber)
	})
}
