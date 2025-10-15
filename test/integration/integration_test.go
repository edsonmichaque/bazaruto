package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/database"
	"github.com/edsonmichaque/bazaruto/internal/handlers"
	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestServer represents a test server instance
type TestServer struct {
	DB     *gorm.DB
	Router chi.Router
	Stores *store.Stores
}

// SetupTestServer creates a test server with in-memory database
func SetupTestServer(t *testing.T) *TestServer {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = database.RunMigrations(db)
	require.NoError(t, err)

	// Create stores
	stores := store.NewStores(db)

	// Create services
	productService := services.NewProductService(stores.Products)
	quoteService := services.NewQuoteService(stores.Quotes)
	policyService := services.NewPolicyService(stores.Policies)
	claimService := services.NewClaimService(stores.Claims, stores.Policies)

	// Create handlers
	productHandler := handlers.NewProductHandler(productService)
	quoteHandler := handlers.NewQuoteHandler(quoteService)
	policyHandler := handlers.NewPolicyHandler(policyService)
	claimHandler := handlers.NewClaimHandler(claimService)

	// Create router
	r := chi.NewRouter()

	// Register routes
	r.Route("/v1", func(r chi.Router) {
		productHandler.RegisterRoutes(r)
		quoteHandler.RegisterRoutes(r)
		policyHandler.RegisterRoutes(r)
		claimHandler.RegisterRoutes(r)
	})

	return &TestServer{
		DB:     db,
		Router: r,
		Stores: stores,
	}
}

// TestProductCRUD tests the complete CRUD operations for products
func TestProductCRUD(t *testing.T) {
	server := SetupTestServer(t)

	// Create a test partner first
	partner := &models.Partner{
		Name:           "Test Partner",
		Description:    "Test partner for integration tests",
		LicenseNumber:  "TEST-001",
		Status:         models.StatusActive,
		CommissionRate: 0.1,
	}
	err := server.DB.Create(partner).Error
	require.NoError(t, err)

	// Test Create Product
	product := &models.Product{
		Name:           "Test Product",
		Description:    "Test product for integration tests",
		Category:       "health",
		PartnerID:      partner.ID,
		BasePrice:      100.0,
		Currency:       models.CurrencyUSD,
		CoverageAmount: 10000.0,
		CoveragePeriod: 365,
		Deductible:     500.0,
		Status:         models.StatusActive,
		EffectiveDate:  time.Now(),
	}

	productJSON, err := json.Marshal(product)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/v1/products", bytes.NewBuffer(productJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdProduct models.Product
	err = json.Unmarshal(w.Body.Bytes(), &createdProduct)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, createdProduct.ID)
	assert.Equal(t, product.Name, createdProduct.Name)

	// Test Get Product
	req = httptest.NewRequest("GET", fmt.Sprintf("/v1/products/%s", createdProduct.ID), nil)
	w = httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var retrievedProduct models.Product
	err = json.Unmarshal(w.Body.Bytes(), &retrievedProduct)
	require.NoError(t, err)
	assert.Equal(t, createdProduct.ID, retrievedProduct.ID)

	// Test List Products
	req = httptest.NewRequest("GET", "/v1/products", nil)
	w = httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "data")

	// Test Update Product
	createdProduct.Name = "Updated Test Product"
	updatedJSON, err := json.Marshal(createdProduct)
	require.NoError(t, err)

	req = httptest.NewRequest("PUT", fmt.Sprintf("/v1/products/%s", createdProduct.ID), bytes.NewBuffer(updatedJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test Delete Product
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/v1/products/%s", createdProduct.ID), nil)
	w = httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// TestQuoteWorkflow tests the quote creation and management workflow
func TestQuoteWorkflow(t *testing.T) {
	server := SetupTestServer(t)

	// Create test data
	partner := &models.Partner{
		Name:           "Test Partner",
		LicenseNumber:  "TEST-002",
		Status:         models.StatusActive,
		CommissionRate: 0.1,
	}
	err := server.DB.Create(partner).Error
	require.NoError(t, err)

	product := &models.Product{
		Name:           "Test Product",
		Category:       "health",
		PartnerID:      partner.ID,
		BasePrice:      100.0,
		Currency:       models.CurrencyUSD,
		CoverageAmount: 10000.0,
		CoveragePeriod: 365,
		Status:         models.StatusActive,
		EffectiveDate:  time.Now(),
	}
	err = server.DB.Create(product).Error
	require.NoError(t, err)

	user := &models.User{
		Email:        "test@example.com",
		FullName:     "Test User",
		PasswordHash: "test-hash",
		Status:       models.StatusActive,
	}
	err = server.DB.Create(user).Error
	require.NoError(t, err)

	// Test Create Quote
	quote := &models.Quote{
		ProductID:  product.ID,
		UserID:     user.ID,
		BasePrice:  100.0,
		FinalPrice: 100.0,
		Currency:   models.CurrencyUSD,
		Status:     models.QuoteStatusPending,
		ValidUntil: time.Now().Add(24 * time.Hour),
	}

	quoteJSON, err := json.Marshal(quote)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/v1/quotes", bytes.NewBuffer(quoteJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdQuote models.Quote
	err = json.Unmarshal(w.Body.Bytes(), &createdQuote)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, createdQuote.ID)
	assert.NotEmpty(t, createdQuote.QuoteNumber)
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	server := SetupTestServer(t)

	// Add health endpoint
	server.Router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":    "healthy",
			"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		})
	})

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Set test environment variables
	_ = os.Setenv("BAZARUTO_LOG_LEVEL", "error")
	_ = os.Setenv("BAZARUTO_LOG_FORMAT", "json")

	// Run tests
	code := m.Run()

	// Cleanup
	os.Exit(code)
}
