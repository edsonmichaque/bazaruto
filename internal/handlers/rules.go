package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/edsonmichaque/bazaruto/internal/config"
	"github.com/edsonmichaque/bazaruto/internal/logger"
)

// RulesHandler handles business rules configuration endpoints
type RulesHandler struct {
	configManager *config.Manager
	logger        *logger.Logger
}

// NewRulesHandler creates a new rules handler
func NewRulesHandler(configManager *config.Manager, logger *logger.Logger) *RulesHandler {
	return &RulesHandler{
		configManager: configManager,
		logger:        logger,
	}
}

// GetRules handles GET /v1/rules
func (h *RulesHandler) GetRules(w http.ResponseWriter, r *http.Request) {
	// Get current business rules configuration
	rules := h.configManager.GetConfig()

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Write response
	if err := json.NewEncoder(w).Encode(rules); err != nil {
		h.logger.Error("Failed to encode rules response", zap.Error(err))
		_ = writeInternalError(w, err)
		return
	}
}

// GetRuleSection handles GET /v1/rules/{section}
func (h *RulesHandler) GetRuleSection(w http.ResponseWriter, r *http.Request) {
	// Get section name from URL parameter
	section := chi.URLParam(r, "section")
	if section == "" {
		_ = writeValidationError(w, "section parameter is required")
		return
	}

	// Get current business rules configuration
	rules := h.configManager.GetConfig()

	// Extract specific section
	var sectionData interface{}
	switch section {
	case "fraud_detection":
		sectionData = rules.FraudDetection
	case "risk_assessment":
		sectionData = rules.RiskAssessment
	case "pricing":
		sectionData = rules.Pricing
	case "underwriting":
		sectionData = rules.Underwriting
	case "commission":
		sectionData = rules.Commission
	case "compliance":
		sectionData = rules.Compliance
	case "policy_lifecycle":
		sectionData = rules.PolicyLifecycle
	case "claim_processing":
		sectionData = rules.ClaimProcessing
	default:
		_ = writeValidationError(w, "invalid section: "+section)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Write response
	if err := json.NewEncoder(w).Encode(sectionData); err != nil {
		h.logger.Error("Failed to encode rule section response", zap.Error(err))
		_ = writeInternalError(w, err)
		return
	}
}

// UpdateRules handles PUT /v1/rules
func (h *RulesHandler) UpdateRules(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var newRules config.BusinessRulesConfig
	if err := json.NewDecoder(r.Body).Decode(&newRules); err != nil {
		_ = writeValidationError(w, "invalid JSON: "+err.Error())
		return
	}

	// Validate the configuration
	if err := h.configManager.ValidateConfig(&newRules); err != nil {
		_ = writeValidationError(w, "invalid configuration: "+err.Error())
		return
	}

	// Update configuration
	if err := h.configManager.UpdateConfig(r.Context(), &newRules); err != nil {
		h.logger.Error("Failed to update rules configuration", zap.Error(err))
		_ = writeInternalError(w, err)
		return
	}

	// Log the update
	h.logger.Info("Business rules configuration updated",
		zap.String("user_agent", r.UserAgent()),
		zap.String("remote_addr", r.RemoteAddr))

	// Return updated configuration
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(newRules); err != nil {
		h.logger.Error("Failed to encode updated rules response", zap.Error(err))
		_ = writeInternalError(w, err)
		return
	}
}

// UpdateRuleSection handles PUT /v1/rules/{section}
func (h *RulesHandler) UpdateRuleSection(w http.ResponseWriter, r *http.Request) {
	// Get section name from URL parameter
	section := chi.URLParam(r, "section")
	if section == "" {
		_ = writeValidationError(w, "section parameter is required")
		return
	}

	// Get current business rules configuration
	currentRules := h.configManager.GetConfig()

	// Parse request body based on section type
	var sectionData interface{}
	var err error

	switch section {
	case "fraud_detection":
		var fraudConfig config.FraudDetectionConfig
		err = json.NewDecoder(r.Body).Decode(&fraudConfig)
		sectionData = fraudConfig
		currentRules.FraudDetection = fraudConfig
	case "risk_assessment":
		var riskConfig config.RiskAssessmentConfig
		err = json.NewDecoder(r.Body).Decode(&riskConfig)
		sectionData = riskConfig
		currentRules.RiskAssessment = riskConfig
	case "pricing":
		var pricingConfig config.PricingConfig
		err = json.NewDecoder(r.Body).Decode(&pricingConfig)
		sectionData = pricingConfig
		currentRules.Pricing = pricingConfig
	case "underwriting":
		var underwritingConfig config.UnderwritingConfig
		err = json.NewDecoder(r.Body).Decode(&underwritingConfig)
		sectionData = underwritingConfig
		currentRules.Underwriting = underwritingConfig
	case "commission":
		var commissionConfig config.CommissionConfig
		err = json.NewDecoder(r.Body).Decode(&commissionConfig)
		sectionData = commissionConfig
		currentRules.Commission = commissionConfig
	case "compliance":
		var complianceConfig config.ComplianceConfig
		err = json.NewDecoder(r.Body).Decode(&complianceConfig)
		sectionData = complianceConfig
		currentRules.Compliance = complianceConfig
	case "policy_lifecycle":
		var lifecycleConfig config.PolicyLifecycleConfig
		err = json.NewDecoder(r.Body).Decode(&lifecycleConfig)
		sectionData = lifecycleConfig
		currentRules.PolicyLifecycle = lifecycleConfig
	case "claim_processing":
		var claimConfig config.ClaimProcessingConfig
		err = json.NewDecoder(r.Body).Decode(&claimConfig)
		sectionData = claimConfig
		currentRules.ClaimProcessing = claimConfig
	default:
		_ = writeValidationError(w, "invalid section: "+section)
		return
	}

	if err != nil {
		_ = writeValidationError(w, "invalid JSON: "+err.Error())
		return
	}

	// Validate the updated configuration
	if err := h.configManager.ValidateConfig(currentRules); err != nil {
		_ = writeValidationError(w, "invalid configuration: "+err.Error())
		return
	}

	// Update configuration
	if err := h.configManager.UpdateConfig(r.Context(), currentRules); err != nil {
		h.logger.Error("Failed to update rule section", zap.Error(err), zap.String("section", section))
		_ = writeInternalError(w, err)
		return
	}

	// Log the update
	h.logger.Info("Business rule section updated",
		zap.String("section", section),
		zap.String("user_agent", r.UserAgent()),
		zap.String("remote_addr", r.RemoteAddr))

	// Return updated section
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sectionData); err != nil {
		h.logger.Error("Failed to encode updated rule section response", zap.Error(err))
		_ = writeInternalError(w, err)
		return
	}
}

// GetRulesVersion handles GET /v1/rules/version
func (h *RulesHandler) GetRulesVersion(w http.ResponseWriter, r *http.Request) {
	// Get configuration metadata
	metadata := h.configManager.GetMetadata()

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Write response
	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		h.logger.Error("Failed to encode rules version response", zap.Error(err))
		_ = writeInternalError(w, err)
		return
	}
}

// ReloadRules handles POST /v1/rules/reload
func (h *RulesHandler) ReloadRules(w http.ResponseWriter, r *http.Request) {
	// Reload configuration from file
	if err := h.configManager.LoadConfig(r.Context()); err != nil {
		h.logger.Error("Failed to reload rules configuration", zap.Error(err))
		_ = writeInternalError(w, err)
		return
	}

	// Log the reload
	h.logger.Info("Business rules configuration reloaded from file",
		zap.String("user_agent", r.UserAgent()),
		zap.String("remote_addr", r.RemoteAddr))

	// Return success response
	response := map[string]interface{}{
		"message":   "Configuration reloaded successfully",
		"timestamp": h.configManager.GetMetadata().LastUpdated,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode reload response", zap.Error(err))
		_ = writeInternalError(w, err)
		return
	}
}

// RegisterRoutes registers rules routes with the router
func (h *RulesHandler) RegisterRoutes(r chi.Router) {
	// Rules endpoints
	r.Get("/rules", h.GetRules)
	r.Put("/rules", h.UpdateRules)
	r.Get("/rules/version", h.GetRulesVersion)
	r.Post("/rules/reload", h.ReloadRules)

	// Section-specific endpoints
	r.Get("/rules/{section}", h.GetRuleSection)
	r.Put("/rules/{section}", h.UpdateRuleSection)
}
