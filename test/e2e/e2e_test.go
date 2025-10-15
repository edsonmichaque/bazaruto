package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// E2ETestClient represents a client for end-to-end testing
type E2ETestClient struct {
	BaseURL string
	Client  *http.Client
}

// NewE2ETestClient creates a new E2E test client
func NewE2ETestClient(baseURL string) *E2ETestClient {
	return &E2ETestClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HealthCheck tests the health endpoint
func (c *E2ETestClient) HealthCheck(t *testing.T) {
	resp, err := c.Client.Get(c.BaseURL + "/healthz")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

// CreatePartner creates a test partner
func (c *E2ETestClient) CreatePartner(t *testing.T) *models.Partner {
	partner := &models.Partner{
		Name:           "E2E Test Partner",
		Description:    "Partner for E2E testing",
		LicenseNumber:  "E2E-001",
		Status:         models.StatusActive,
		CommissionRate: 0.1,
	}

	partnerJSON, err := json.Marshal(partner)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/partners", bytes.NewBuffer(partnerJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdPartner models.Partner
	err = json.NewDecoder(resp.Body).Decode(&createdPartner)
	require.NoError(t, err)

	return &createdPartner
}

// CreateProduct creates a test product
func (c *E2ETestClient) CreateProduct(t *testing.T, partnerID string) *models.Product {
	partnerUUID, err := uuid.Parse(partnerID)
	require.NoError(t, err)

	product := &models.Product{
		Name:           "E2E Test Product",
		Description:    "Product for E2E testing",
		Category:       "health",
		PartnerID:      partnerUUID,
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

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/products", bytes.NewBuffer(productJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdProduct models.Product
	err = json.NewDecoder(resp.Body).Decode(&createdProduct)
	require.NoError(t, err)

	return &createdProduct
}

// CreateUser creates a test user
func (c *E2ETestClient) CreateUser(t *testing.T) *models.User {
	user := &models.User{
		Email:        "e2e@example.com",
		FullName:     "E2E Test User",
		PasswordHash: "test-hash",
		Status:       models.StatusActive,
	}

	userJSON, err := json.Marshal(user)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/users", bytes.NewBuffer(userJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdUser models.User
	err = json.NewDecoder(resp.Body).Decode(&createdUser)
	require.NoError(t, err)

	return &createdUser
}

// CreateQuote creates a test quote
func (c *E2ETestClient) CreateQuote(t *testing.T, productID, userID string) *models.Quote {
	productUUID, err := uuid.Parse(productID)
	require.NoError(t, err)
	userUUID, err := uuid.Parse(userID)
	require.NoError(t, err)

	quote := &models.Quote{
		ProductID:  productUUID,
		UserID:     userUUID,
		BasePrice:  100.0,
		FinalPrice: 100.0,
		Currency:   models.CurrencyUSD,
		Status:     models.QuoteStatusPending,
		ValidUntil: time.Now().Add(24 * time.Hour),
	}

	quoteJSON, err := json.Marshal(quote)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/quotes", bytes.NewBuffer(quoteJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdQuote models.Quote
	err = json.NewDecoder(resp.Body).Decode(&createdQuote)
	require.NoError(t, err)

	return &createdQuote
}

// TestE2EHealthCheck tests the health endpoint in a real environment
func TestE2EHealthCheck(t *testing.T) {
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := NewE2ETestClient(baseURL)
	client.HealthCheck(t)
}

// TestE2EProductWorkflow tests the complete product workflow
func TestE2EProductWorkflow(t *testing.T) {
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := NewE2ETestClient(baseURL)

	// Test health check first
	client.HealthCheck(t)

	// Create partner
	partner := client.CreatePartner(t)
	require.NotNil(t, partner)

	// Create product
	product := client.CreateProduct(t, partner.ID.String())
	require.NotNil(t, product)

	// Verify product was created
	resp, err := client.Client.Get(fmt.Sprintf("%s/v1/products/%s", baseURL, product.ID))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedProduct models.Product
	err = json.NewDecoder(resp.Body).Decode(&retrievedProduct)
	require.NoError(t, err)
	assert.Equal(t, product.ID, retrievedProduct.ID)
}

// TestE2EQuoteWorkflow tests the complete quote workflow
func TestE2EQuoteWorkflow(t *testing.T) {
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := NewE2ETestClient(baseURL)

	// Create test data
	partner := client.CreatePartner(t)
	product := client.CreateProduct(t, partner.ID.String())
	user := client.CreateUser(t)

	// Create quote
	quote := client.CreateQuote(t, product.ID.String(), user.ID.String())
	require.NotNil(t, quote)

	// Verify quote was created
	resp, err := client.Client.Get(fmt.Sprintf("%s/v1/quotes/%s", baseURL, quote.ID))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedQuote models.Quote
	err = json.NewDecoder(resp.Body).Decode(&retrievedQuote)
	require.NoError(t, err)
	assert.Equal(t, quote.ID, retrievedQuote.ID)
}

// TestE2ERateLimiting tests rate limiting functionality
func TestE2ERateLimiting(t *testing.T) {
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := NewE2ETestClient(baseURL)

	// Make multiple requests quickly to trigger rate limiting
	for i := 0; i < 10; i++ {
		resp, err := client.Client.Get(baseURL + "/healthz")
		require.NoError(t, err)
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			t.Logf("Rate limiting triggered after %d requests", i+1)
			return
		}
	}

	t.Log("Rate limiting not triggered within 10 requests")
}

// TestMain sets up the E2E test environment
func TestMain(m *testing.M) {
	// Set test environment variables
	_ = os.Setenv("BAZARUTO_LOG_LEVEL", "error")
	_ = os.Setenv("BAZARUTO_LOG_FORMAT", "json")

	// Wait for service to be ready
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for i := 0; i < 30; i++ {
		resp, err := client.Get(baseURL + "/healthz")
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			break
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.Exit(code)
}
