package store

import (
	"gorm.io/gorm"
)

// Stores aggregates all store interfaces for dependency injection.
type Stores struct {
	Users         UserStore
	Partners      PartnerStore
	Products      ProductStore
	Quotes        QuoteStore
	Policies      PolicyStore
	Claims        ClaimStore
	Subscriptions SubscriptionStore
	Payments      PaymentStore
	Invoices      InvoiceStore
	Beneficiaries BeneficiaryStore
	Coverages     CoverageStore
	Webhooks      WebhookStore
}

// NewStores creates a new Stores instance with all store implementations.
func NewStores(db *gorm.DB) *Stores {
	return &Stores{
		Users:         NewUserStore(db),
		Partners:      NewPartnerStore(db),
		Products:      NewProductStore(db),
		Quotes:        NewQuoteStore(db),
		Policies:      NewPolicyStore(db),
		Claims:        NewClaimStore(db),
		Subscriptions: NewSubscriptionStore(db),
		Payments:      NewPaymentStore(db),
		Invoices:      NewInvoiceStore(db),
		Beneficiaries: NewBeneficiaryStore(db),
		Coverages:     NewCoverageStore(db),
		Webhooks:      NewWebhookStore(db),
	}
}
