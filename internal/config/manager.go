package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/logger"
	"go.uber.org/zap"
)

// ConfigManager manages business rule configurations.
type ConfigManager struct {
	logger      *logger.Logger
	config      *BusinessRulesConfig
	lastUpdated time.Time
	mutex       sync.RWMutex
	configPath  string
}

// NewConfigManager creates a new configuration manager.
func NewConfigManager(logger *logger.Logger, configPath string) *ConfigManager {
	if configPath == "" {
		configPath = "config/business_rules.json"
	}

	return &ConfigManager{
		logger:     logger,
		configPath: configPath,
	}
}

// LoadConfig loads the business rules configuration from file.
func (m *ConfigManager) LoadConfig(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Try to load from file first
	if err := m.loadFromFile(); err != nil {
		m.logger.Warn("Failed to load config from file, using defaults", zap.Error(err))
		m.config = m.getDefaultConfig()
	} else {
		m.lastUpdated = time.Now()
	}

	m.logger.Info("Configuration loaded successfully",
		zap.String("config_path", m.configPath),
		zap.Time("last_updated", m.lastUpdated))

	return nil
}

// loadFromFile loads configuration from a JSON file.
func (m *ConfigManager) loadFromFile() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config BusinessRulesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	m.config = &config
	return nil
}

// SaveConfig saves the current configuration to file.
func (m *ConfigManager) SaveConfig(ctx context.Context) error {
	m.mutex.RLock()
	config := m.config
	m.mutex.RUnlock()

	if config == nil {
		return fmt.Errorf("no configuration to save")
	}

	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	m.mutex.Lock()
	m.lastUpdated = time.Now()
	m.mutex.Unlock()

	m.logger.Info("Configuration saved successfully",
		zap.String("config_path", m.configPath))

	return nil
}

// GetConfig returns the current business rules configuration.
func (m *ConfigManager) GetConfig() *BusinessRulesConfig {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.config == nil {
		return m.getDefaultConfig()
	}

	// Return a copy to prevent external modifications
	configCopy := *m.config
	return &configCopy
}

// UpdateConfig updates the business rules configuration.
func (m *ConfigManager) UpdateConfig(ctx context.Context, config *BusinessRulesConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Validate the configuration
	if err := m.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m.mutex.Lock()
	m.config = config
	m.lastUpdated = time.Now()
	m.mutex.Unlock()

	// Save to file
	if err := m.SaveConfig(ctx); err != nil {
		m.logger.Error("Failed to save updated configuration", zap.Error(err))
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	m.logger.Info("Configuration updated successfully")
	return nil
}

// RefreshConfig refreshes the configuration from file.
func (m *ConfigManager) RefreshConfig(ctx context.Context) error {
	return m.LoadConfig(ctx)
}

// GetLastUpdated returns when the configuration was last updated.
func (m *ConfigManager) GetLastUpdated() time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.lastUpdated
}

// validateConfig validates the business rules configuration.
func (m *ConfigManager) validateConfig(config *BusinessRulesConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Basic validation - can be extended as needed
	return nil
}

// getDefaultConfig returns the default business rules configuration.
func (m *ConfigManager) getDefaultConfig() *BusinessRulesConfig {
	return &BusinessRulesConfig{
		FraudDetection: FraudDetectionConfig{
			Enabled: true,
			Version: "1.0",
			RiskThresholds: RiskThresholds{
				Low:      30.0,
				Medium:   60.0,
				High:     80.0,
				Critical: 90.0,
			},
			FactorWeights: map[string]float64{
				"claim_frequency": 0.3,
				"claim_amount":    0.25,
				"geographic_risk": 0.2,
				"payment_history": 0.15,
				"policy_duration": 0.1,
			},
			TimingRules: TimingRules{
				NewAccountThreshold:     6 * 30 * 24 * time.Hour, // 6 months
				PolicyStartThreshold:    7 * 24 * time.Hour,      // 7 days
				ReportingDelayThreshold: 30 * 24 * time.Hour,     // 30 days
				WeekendMultiplier:       1.2,
				BusinessHoursMultiplier: 0.9,
			},
			AmountRules: AmountRules{
				HighValueThreshold: 10000.0,
			},
			DocumentRules: DocumentRules{
				MinDocumentCount:      2,
				MinFileSize:           1024,             // 1KB
				MaxFileSize:           10 * 1024 * 1024, // 10MB
				RequiredDocumentTypes: []string{"id", "proof_of_address", "financial_statement"},
			},
			GeographicRules: GeographicRules{
				HighRiskCountries: []string{"AF", "IR", "KP", "SY"},
				HighRiskRegions:   []string{"middle_east", "africa"},
				CountryRiskMultiplier: map[string]float64{
					"AF": 2.0,
					"IR": 2.5,
					"KP": 3.0,
					"SY": 2.0,
				},
			},
			BehavioralRules: BehavioralRules{
				MinDescriptionLength:     50,
				MaxDescriptionLength:     1000,
				DescriptionLengthPenalty: 10.0,
			},
			ConfidenceThresholds: ConfidenceThresholds{
				Low:      0.3,
				Medium:   0.5,
				High:     0.7,
				VeryHigh: 0.9,
			},
			AutoReviewThresholds: AutoReviewThresholds{
				ScoreThreshold:      70.0,
				CriticalFactorCount: 1,
				HighSeverityCount:   2,
			},
		},
		RiskAssessment: RiskAssessmentConfig{
			Enabled: true,
			Version: "1.0",
			FactorWeights: map[string]float64{
				"age":            0.2,
				"occupation":     0.25,
				"health_status":  0.3,
				"lifestyle":      0.15,
				"family_history": 0.1,
			},
		},
		Pricing: PricingConfig{
			Enabled: true,
			Version: "1.0",
			BaseRates: map[string]float64{
				"life_insurance":       50.0,
				"health_insurance":     200.0,
				"auto_insurance":       100.0,
				"home_insurance":       150.0,
				"travel_insurance":     25.0,
				"disability_insurance": 75.0,
			},
		},
		Underwriting: UnderwritingConfig{
			Enabled: true,
			Version: "1.0",
		},
		Commission: CommissionConfig{
			Enabled: true,
			Version: "1.0",
		},
		Compliance: ComplianceConfig{
			Enabled: true,
			Version: "1.0",
		},
		PolicyLifecycle: PolicyLifecycleConfig{
			Enabled: true,
			Version: "1.0",
		},
		ClaimProcessing: ClaimProcessingConfig{
			Enabled: true,
			Version: "1.0",
		},
	}
}
