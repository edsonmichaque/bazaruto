package handlers

import (
	"net/http"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PolicyHandler handles HTTP requests for policies.
type PolicyHandler struct {
	service *services.PolicyService
}

// NewPolicyHandler creates a new PolicyHandler.
func NewPolicyHandler(service *services.PolicyService) *PolicyHandler {
	return &PolicyHandler{
		service: service,
	}
}

// ListPolicies handles GET /v1/policies.
func (h *PolicyHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	pagination, err := parsePagination(r)
	if err != nil {
		_ = writeValidationError(w, err.Error())
		return
	}

	// Create policy list options
	opts := models.NewPolicyListOptions()
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

	// Get policies
	policies, err := h.service.ListPolicies(r.Context(), opts)
	if err != nil {
		_ = writeInternalError(w, err)
		return
	}

	// Get total count
	total, err := h.service.CountPolicies(r.Context(), opts)
	if err != nil {
		_ = writeInternalError(w, err)
		return
	}

	// Write page-based paginated response
	if err := writePagePaginatedResponse(w, r, policies, total, opts.Page, opts.PerPage); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// GetPolicy handles GET /v1/policies/{id}.
func (h *PolicyHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	// Parse policy ID
	policyID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid policy ID")
		return
	}

	// Get policy
	policy, err := h.service.GetPolicy(r.Context(), policyID)
	if err != nil {
		if err.Error() == "policy not found" {
			_ = writeNotFound(w, "Policy")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	_ = WriteJSONIgnoreError(w, http.StatusOK, policy)
}

// GetPolicyByNumber handles GET /v1/policies/number/{number}.
func (h *PolicyHandler) GetPolicyByNumber(w http.ResponseWriter, r *http.Request) {
	// Parse policy number
	policyNumber := param(r, "number")
	if policyNumber == "" {
		_ = writeValidationError(w, "Policy number is required")
		return
	}

	// Get policy
	policy, err := h.service.GetPolicyByNumber(r.Context(), policyNumber)
	if err != nil {
		if err.Error() == "policy not found" {
			_ = writeNotFound(w, "Policy")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	_ = WriteJSONIgnoreError(w, http.StatusOK, policy)
}

// CreatePolicy handles POST /v1/policies.
func (h *PolicyHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var policy models.Policy
	if err := parseJSON(r, &policy); err != nil {
		_ = writeValidationError(w, "Invalid JSON body")
		return
	}

	// Create policy
	if err := h.service.CreatePolicy(r.Context(), &policy); err != nil {
		_ = writeValidationError(w, err.Error())
		return
	}

	// Write response
	_ = WriteJSONIgnoreError(w, http.StatusCreated, policy)
}

// UpdatePolicy handles PUT /v1/policies/{id}.
func (h *PolicyHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	// Parse policy ID
	policyID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid policy ID")
		return
	}

	// Parse request body
	var policy models.Policy
	if err := parseJSON(r, &policy); err != nil {
		_ = writeValidationError(w, "Invalid JSON body")
		return
	}

	// Set ID from URL
	policy.ID = policyID

	// Update policy
	if err := h.service.UpdatePolicy(r.Context(), &policy); err != nil {
		if err.Error() == "policy not found" {
			_ = writeNotFound(w, "Policy")
			return
		}
		_ = writeValidationError(w, err.Error())
		return
	}

	// Write response
	_ = WriteJSONIgnoreError(w, http.StatusOK, policy)
}

// DeletePolicy handles DELETE /v1/policies/{id}.
func (h *PolicyHandler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	// Parse policy ID
	policyID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid policy ID")
		return
	}

	// Delete policy
	if err := h.service.DeletePolicy(r.Context(), policyID); err != nil {
		if err.Error() == "policy not found" {
			_ = writeNotFound(w, "Policy")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes registers policy routes with the router.
func (h *PolicyHandler) RegisterRoutes(r chi.Router) {
	r.Route("/policies", func(r chi.Router) {
		r.Get("/", h.ListPolicies)
		r.Post("/", h.CreatePolicy)
		r.Get("/{id}", h.GetPolicy)
		r.Put("/{id}", h.UpdatePolicy)
		r.Delete("/{id}", h.DeletePolicy)
		r.Get("/number/{number}", h.GetPolicyByNumber)
	})
}
