package handlers

import (
	"net/http"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ProductHandler handles HTTP requests for products.
type ProductHandler struct {
	service *services.ProductService
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

// ListProducts handles GET /v1/products.
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	pagination, err := parsePagination(r)
	if err != nil {
		_ = writeValidationError(w, err.Error())
		return
	}

	// Create product list options
	opts := models.NewProductListOptions()
	opts.Page = pagination.Page
	opts.PerPage = pagination.PerPage

	// Parse filter parameters
	if partnerIDStr := queryParam(r, "partner_id"); partnerIDStr != "" {
		if id, err := uuid.Parse(partnerIDStr); err == nil {
			opts.PartnerID = &id
		}
	}

	opts.Category = queryParam(r, "category")
	opts.Currency = queryParam(r, "currency")

	// Get products
	products, err := h.service.ListProducts(r.Context(), opts)
	if err != nil {
		_ = writeInternalError(w, err)
		return
	}

	// Get total count
	total, err := h.service.CountProducts(r.Context(), opts)
	if err != nil {
		_ = writeInternalError(w, err)
		return
	}

	// Write page-based paginated response
	if err := writePagePaginatedResponse(w, r, products, total, opts.Page, opts.PerPage); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// GetProduct handles GET /v1/products/{id}.
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	// Parse product ID
	productID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid product ID")
		return
	}

	// Get product
	product, err := h.service.GetProduct(r.Context(), productID)
	if err != nil {
		if err.Error() == "product not found" {
			_ = writeNotFound(w, "Product")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	if err := WriteJSONIgnoreError(w, http.StatusOK, product); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// CreateProduct handles POST /v1/products.
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var product models.Product
	if err := parseJSON(r, &product); err != nil {
		_ = writeValidationError(w, "Invalid JSON body")
		return
	}

	// Create product
	if err := h.service.CreateProduct(r.Context(), &product); err != nil {
		_ = writeValidationError(w, err.Error())
		return
	}

	// Write response
	if err := WriteJSONIgnoreError(w, http.StatusCreated, product); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// UpdateProduct handles PUT /v1/products/{id}.
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	// Parse product ID
	productID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid product ID")
		return
	}

	// Parse request body
	var product models.Product
	if err := parseJSON(r, &product); err != nil {
		_ = writeValidationError(w, "Invalid JSON body")
		return
	}

	// Set ID from URL
	product.ID = productID

	// Update product
	if err := h.service.UpdateProduct(r.Context(), &product); err != nil {
		if err.Error() == "product not found" {
			_ = writeNotFound(w, "Product")
			return
		}
		_ = writeValidationError(w, err.Error())
		return
	}

	// Write response
	if err := WriteJSONIgnoreError(w, http.StatusOK, product); err != nil {
		_ = writeInternalError(w, err)
		return
	}
}

// DeleteProduct handles DELETE /v1/products/{id}.
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// Parse product ID
	productID, err := paramUUID(r, "id")
	if err != nil {
		_ = writeValidationError(w, "Invalid product ID")
		return
	}

	// Delete product
	if err := h.service.DeleteProduct(r.Context(), productID); err != nil {
		if err.Error() == "product not found" {
			_ = writeNotFound(w, "Product")
			return
		}
		_ = writeInternalError(w, err)
		return
	}

	// Write response
	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes registers product routes with the router.
func (h *ProductHandler) RegisterRoutes(r chi.Router) {
	r.Route("/products", func(r chi.Router) {
		r.Get("/", h.ListProducts)
		r.Post("/", h.CreateProduct)
		r.Get("/{id}", h.GetProduct)
		r.Put("/{id}", h.UpdateProduct)
		r.Delete("/{id}", h.DeleteProduct)
	})
}
