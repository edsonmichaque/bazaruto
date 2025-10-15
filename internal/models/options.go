package models

import (
	"time"

	"github.com/google/uuid"
)

// ListOptions provides standardized pagination and filtering options for list operations
// Uses GitHub-like pagination with page/per_page instead of limit/offset
type ListOptions struct {
	// GitHub-style pagination
	Page    int `json:"page" form:"page"`         // Page number (1-based)
	PerPage int `json:"per_page" form:"per_page"` // Items per page

	// Sorting
	SortBy    string `json:"sort" form:"sort"`   // Field to sort by (GitHub style)
	SortOrder string `json:"order" form:"order"` // "asc" or "desc" (GitHub style)

	// Common filters
	Status string `json:"status" form:"status"`

	// Date range filters
	CreatedAfter  *time.Time `json:"created_after" form:"created_after"`
	CreatedBefore *time.Time `json:"created_before" form:"created_before"`
	UpdatedAfter  *time.Time `json:"updated_after" form:"updated_after"`
	UpdatedBefore *time.Time `json:"updated_before" form:"updated_before"`

	// Search
	Search string `json:"search" form:"search"` // General search term

	// Additional filters (can be extended per entity)
	Filters map[string]interface{} `json:"filters" form:"filters"`
}

// DefaultListOptions returns sensible defaults for list operations
func DefaultListOptions() *ListOptions {
	return &ListOptions{
		Page:      1,
		PerPage:   30, // GitHub default
		SortBy:    "created_at",
		SortOrder: "desc",
		Filters:   make(map[string]interface{}),
	}
}

// Validate validates and normalizes the list options
func (opts *ListOptions) Validate() error {
	// Set defaults if not provided
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PerPage <= 0 {
		opts.PerPage = 30
	}
	if opts.PerPage > 100 {
		opts.PerPage = 100 // GitHub's max per_page
	}
	if opts.SortBy == "" {
		opts.SortBy = "created_at"
	}
	if opts.SortOrder == "" {
		opts.SortOrder = "desc"
	}
	if opts.SortOrder != "asc" && opts.SortOrder != "desc" {
		opts.SortOrder = "desc"
	}
	if opts.Filters == nil {
		opts.Filters = make(map[string]interface{})
	}

	return nil
}

// GetSortClause returns the SQL ORDER BY clause
func (opts *ListOptions) GetSortClause() string {
	return opts.SortBy + " " + opts.SortOrder
}

// GetLimit returns the limit value (for SQL queries)
func (opts *ListOptions) GetLimit() int {
	return opts.PerPage
}

// GetOffset returns the offset value (for SQL queries)
func (opts *ListOptions) GetOffset() int {
	return (opts.Page - 1) * opts.PerPage
}

// GetPage returns the current page number
func (opts *ListOptions) GetPage() int {
	if opts.Page <= 0 {
		return 1
	}
	return opts.Page
}

// GetPerPage returns the items per page
func (opts *ListOptions) GetPerPage() int {
	if opts.PerPage <= 0 {
		return 30
	}
	return opts.PerPage
}

// HasDateFilter returns true if any date filters are set
func (opts *ListOptions) HasDateFilter() bool {
	return opts.CreatedAfter != nil || opts.CreatedBefore != nil ||
		opts.UpdatedAfter != nil || opts.UpdatedBefore != nil
}

// HasSearch returns true if search term is provided
func (opts *ListOptions) HasSearch() bool {
	return opts.Search != ""
}

// HasStatus returns true if status filter is provided
func (opts *ListOptions) HasStatus() bool {
	return opts.Status != ""
}

// HasFilter returns true if a specific filter key exists
func (opts *ListOptions) HasFilter(key string) bool {
	_, exists := opts.Filters[key]
	return exists
}

// GetFilter returns the value for a specific filter key
func (opts *ListOptions) GetFilter(key string) (interface{}, bool) {
	value, exists := opts.Filters[key]
	return value, exists
}

// SetFilter sets a filter value
func (opts *ListOptions) SetFilter(key string, value interface{}) {
	if opts.Filters == nil {
		opts.Filters = make(map[string]interface{})
	}
	opts.Filters[key] = value
}

// RemoveFilter removes a filter
func (opts *ListOptions) RemoveFilter(key string) {
	if opts.Filters != nil {
		delete(opts.Filters, key)
	}
}

// ListResponse provides a standardized response structure for list operations
// GitHub-style pagination response
type ListResponse[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total_count"`
	Page     int   `json:"page"`
	PerPage  int   `json:"per_page"`
	HasMore  bool  `json:"has_more"`
	NextPage *int  `json:"next_page,omitempty"`
	PrevPage *int  `json:"prev_page,omitempty"`
	LastPage int   `json:"last_page"`
}

// NewListResponse creates a new list response with GitHub-style pagination
func NewListResponse[T any](items []T, total int64, opts *ListOptions) *ListResponse[T] {
	page := opts.GetPage()
	perPage := opts.GetPerPage()
	lastPage := int((total + int64(perPage) - 1) / int64(perPage)) // Ceiling division

	response := &ListResponse[T]{
		Items:    items,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
		LastPage: lastPage,
	}

	// Calculate pagination links
	response.HasMore = page < lastPage
	if response.HasMore {
		nextPage := page + 1
		response.NextPage = &nextPage
	}
	if page > 1 {
		prevPage := page - 1
		response.PrevPage = &prevPage
	}

	return response
}

// Entity-specific list options

// UserListOptions provides filtering options specific to users
type UserListOptions struct {
	*ListOptions
	Role        string `json:"role" form:"role"`
	EmailDomain string `json:"email_domain" form:"email_domain"`
	IsActive    *bool  `json:"is_active" form:"is_active"`
}

// NewUserListOptions creates user-specific list options
func NewUserListOptions() *UserListOptions {
	return &UserListOptions{
		ListOptions: DefaultListOptions(),
	}
}

// QuoteListOptions provides filtering options specific to quotes
type QuoteListOptions struct {
	*ListOptions
	UserID    *uuid.UUID `json:"user_id" form:"user_id"`
	ProductID *uuid.UUID `json:"product_id" form:"product_id"`
	Currency  string     `json:"currency" form:"currency"`
	MinPrice  *float64   `json:"min_price" form:"min_price"`
	MaxPrice  *float64   `json:"max_price" form:"max_price"`
}

// NewQuoteListOptions creates quote-specific list options
func NewQuoteListOptions() *QuoteListOptions {
	return &QuoteListOptions{
		ListOptions: DefaultListOptions(),
	}
}

// PaymentListOptions provides filtering options specific to payments
type PaymentListOptions struct {
	*ListOptions
	UserID         *uuid.UUID `json:"user_id" form:"user_id"`
	PolicyID       *uuid.UUID `json:"policy_id" form:"policy_id"`
	SubscriptionID *uuid.UUID `json:"subscription_id" form:"subscription_id"`
	PaymentMethod  string     `json:"payment_method" form:"payment_method"`
	Currency       string     `json:"currency" form:"currency"`
	MinAmount      *float64   `json:"min_amount" form:"min_amount"`
	MaxAmount      *float64   `json:"max_amount" form:"max_amount"`
}

// NewPaymentListOptions creates payment-specific list options
func NewPaymentListOptions() *PaymentListOptions {
	return &PaymentListOptions{
		ListOptions: DefaultListOptions(),
	}
}

// PolicyListOptions provides filtering options specific to policies
type PolicyListOptions struct {
	*ListOptions
	UserID     *uuid.UUID `json:"user_id" form:"user_id"`
	ProductID  *uuid.UUID `json:"product_id" form:"product_id"`
	QuoteID    *uuid.UUID `json:"quote_id" form:"quote_id"`
	Currency   string     `json:"currency" form:"currency"`
	MinPremium *float64   `json:"min_premium" form:"min_premium"`
	MaxPremium *float64   `json:"max_premium" form:"max_premium"`
}

// NewPolicyListOptions creates policy-specific list options
func NewPolicyListOptions() *PolicyListOptions {
	return &PolicyListOptions{
		ListOptions: DefaultListOptions(),
	}
}

// ClaimListOptions provides filtering options specific to claims
type ClaimListOptions struct {
	*ListOptions
	UserID    *uuid.UUID `json:"user_id" form:"user_id"`
	PolicyID  *uuid.UUID `json:"policy_id" form:"policy_id"`
	ClaimType string     `json:"claim_type" form:"claim_type"`
	Currency  string     `json:"currency" form:"currency"`
	MinAmount *float64   `json:"min_amount" form:"min_amount"`
	MaxAmount *float64   `json:"max_amount" form:"max_amount"`
}

// NewClaimListOptions creates claim-specific list options
func NewClaimListOptions() *ClaimListOptions {
	return &ClaimListOptions{
		ListOptions: DefaultListOptions(),
	}
}

// WebhookListOptions provides filtering options specific to webhook deliveries
type WebhookListOptions struct {
	*ListOptions
	WebhookConfigID *uuid.UUID `json:"webhook_config_id" form:"webhook_config_id"`
	EventType       string     `json:"event_type" form:"event_type"`
	EventID         *uuid.UUID `json:"event_id" form:"event_id"`
	URL             string     `json:"url" form:"url"`
	Method          string     `json:"method" form:"method"`
}

// NewWebhookListOptions creates webhook-specific list options
func NewWebhookListOptions() *WebhookListOptions {
	return &WebhookListOptions{
		ListOptions: DefaultListOptions(),
	}
}

// CustomerListOptions provides filtering options specific to customers
type CustomerListOptions struct {
	*ListOptions
	UserID         *uuid.UUID `json:"user_id" form:"user_id"`
	RiskProfile    string     `json:"risk_profile" form:"risk_profile"`
	CustomerTier   string     `json:"customer_tier" form:"customer_tier"`
	KYCStatus      string     `json:"kyc_status" form:"kyc_status"`
	AMLStatus      string     `json:"aml_status" form:"aml_status"`
	Nationality    string     `json:"nationality" form:"nationality"`
	Occupation     string     `json:"occupation" form:"occupation"`
	MinIncome      *float64   `json:"min_income" form:"min_income"`
	MaxIncome      *float64   `json:"max_income" form:"max_income"`
	MinCreditScore *int       `json:"min_credit_score" form:"min_credit_score"`
	MaxCreditScore *int       `json:"max_credit_score" form:"max_credit_score"`
	HasAddress     *bool      `json:"has_address" form:"has_address"`
	HasDocuments   *bool      `json:"has_documents" form:"has_documents"`
}

// NewCustomerListOptions creates customer-specific list options
func NewCustomerListOptions() *CustomerListOptions {
	return &CustomerListOptions{
		ListOptions: DefaultListOptions(),
	}
}

// ProductListOptions provides filtering options specific to products
type ProductListOptions struct {
	*ListOptions
	PartnerID *uuid.UUID `json:"partner_id" form:"partner_id"`
	Category  string     `json:"category" form:"category"`
	Currency  string     `json:"currency" form:"currency"`
	MinPrice  *float64   `json:"min_price" form:"min_price"`
	MaxPrice  *float64   `json:"max_price" form:"max_price"`
}

// NewProductListOptions creates product-specific list options
func NewProductListOptions() *ProductListOptions {
	return &ProductListOptions{
		ListOptions: DefaultListOptions(),
	}
}
