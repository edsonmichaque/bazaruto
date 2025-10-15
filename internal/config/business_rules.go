package config

import (
	"time"
)

// BusinessRulesConfig holds all business rule configurations.
type BusinessRulesConfig struct {
	FraudDetection    FraudDetectionConfig    `json:"fraud_detection"`
	RiskAssessment    RiskAssessmentConfig    `json:"risk_assessment"`
	Pricing           PricingConfig           `json:"pricing"`
	Underwriting      UnderwritingConfig      `json:"underwriting"`
	Commission        CommissionConfig        `json:"commission"`
	Compliance        ComplianceConfig        `json:"compliance"`
	PolicyLifecycle   PolicyLifecycleConfig   `json:"policy_lifecycle"`
	ClaimProcessing   ClaimProcessingConfig   `json:"claim_processing"`
}

// FraudDetectionConfig holds fraud detection configuration.
type FraudDetectionConfig struct {
	Enabled                bool                    `json:"enabled"`
	Version               string                  `json:"version"`
	RiskThresholds        RiskThresholds          `json:"risk_thresholds"`
	FactorWeights         map[string]float64      `json:"factor_weights"`
	TimingRules           TimingRules             `json:"timing_rules"`
	AmountRules           AmountRules             `json:"amount_rules"`
	DocumentRules         DocumentRules           `json:"document_rules"`
	GeographicRules       GeographicRules         `json:"geographic_rules"`
	BehavioralRules       BehavioralRules         `json:"behavioral_rules"`
	ConfidenceThresholds  ConfidenceThresholds    `json:"confidence_thresholds"`
	AutoReviewThresholds  AutoReviewThresholds    `json:"auto_review_thresholds"`
}

// RiskThresholds defines risk score thresholds.
type RiskThresholds struct {
	Low      float64 `json:"low"`      // 0-40
	Medium   float64 `json:"medium"`   // 40-60
	High     float64 `json:"high"`     // 60-80
	Critical float64 `json:"critical"` // 80-100
}

// TimingRules defines timing-based fraud detection rules.
type TimingRules struct {
	NewAccountThreshold    time.Duration `json:"new_account_threshold"`    // 6 months
	PolicyStartThreshold   time.Duration `json:"policy_start_threshold"`   // 7 days
	ReportingDelayThreshold time.Duration `json:"reporting_delay_threshold"` // 30 days
	WeekendMultiplier      float64       `json:"weekend_multiplier"`       // 1.2
	BusinessHoursMultiplier float64      `json:"business_hours_multiplier"` // 0.9
}

// AmountRules defines amount-based fraud detection rules.
type AmountRules struct {
	HighValueThreshold     float64 `json:"high_value_threshold"`     // $500,000
	VeryHighValueThreshold float64 `json:"very_high_value_threshold"` // $1,000,000
	RoundNumberPenalty     float64 `json:"round_number_penalty"`     // 15 points
	CoverageRatioThreshold float64 `json:"coverage_ratio_threshold"` // 0.95
}

// DocumentRules defines document-based fraud detection rules.
type DocumentRules struct {
	MinDocumentCount       int     `json:"min_document_count"`       // 2
	MinFileSize            int64   `json:"min_file_size"`            // 1KB
	MaxFileSize            int64   `json:"max_file_size"`            // 10MB
	RequiredDocumentTypes  []string `json:"required_document_types"`
}

// GeographicRules defines geographic-based fraud detection rules.
type GeographicRules struct {
	HighRiskCountries      []string `json:"high_risk_countries"`
	HighRiskRegions        []string `json:"high_risk_regions"`
	CountryRiskMultiplier  map[string]float64 `json:"country_risk_multiplier"`
}

// BehavioralRules defines behavioral-based fraud detection rules.
type BehavioralRules struct {
	MinDescriptionLength   int     `json:"min_description_length"`   // 50
	MaxDescriptionLength   int     `json:"max_description_length"`   // 1000
	DescriptionLengthPenalty float64 `json:"description_length_penalty"` // 10 points
}

// ConfidenceThresholds defines confidence level thresholds.
type ConfidenceThresholds struct {
	Low      float64 `json:"low"`      // 0.3
	Medium   float64 `json:"medium"`   // 0.5
	High     float64 `json:"high"`     // 0.7
	VeryHigh float64 `json:"very_high"` // 0.9
}

// AutoReviewThresholds defines thresholds for automatic review requirements.
type AutoReviewThresholds struct {
	ScoreThreshold         float64 `json:"score_threshold"`         // 70
	CriticalFactorCount    int     `json:"critical_factor_count"`   // 1
	HighSeverityCount      int     `json:"high_severity_count"`     // 2
}

// RiskAssessmentConfig holds risk assessment configuration.
type RiskAssessmentConfig struct {
	Enabled                bool                    `json:"enabled"`
	Version               string                  `json:"version"`
	FactorWeights         map[string]float64      `json:"factor_weights"`
	DemographicRules      DemographicRules        `json:"demographic_rules"`
	BehavioralRules       BehavioralAssessmentRules `json:"behavioral_rules"`
	FinancialRules        FinancialRules          `json:"financial_rules"`
	GeographicRules       GeographicAssessmentRules `json:"geographic_rules"`
	ProductRules          ProductRules            `json:"product_rules"`
	HistoricalRules       HistoricalRules         `json:"historical_rules"`
	LifestyleRules        LifestyleRules          `json:"lifestyle_rules"`
	ComplianceRules       ComplianceAssessmentRules `json:"compliance_rules"`
	ApprovalThresholds    ApprovalThresholds      `json:"approval_thresholds"`
	PremiumAdjustments    PremiumAdjustments      `json:"premium_adjustments"`
}

// DemographicRules defines demographic risk assessment rules.
type DemographicRules struct {
	AgeRanges             []AgeRange              `json:"age_ranges"`
	GenderMultipliers     map[string]float64      `json:"gender_multipliers"`
	MaritalStatusMultipliers map[string]float64   `json:"marital_status_multipliers"`
	OccupationMultipliers map[string]float64      `json:"occupation_multipliers"`
}

// AgeRange defines an age range with associated risk multiplier.
type AgeRange struct {
	Min      int     `json:"min"`
	Max      int     `json:"max"`
	Multiplier float64 `json:"multiplier"`
}

// BehavioralAssessmentRules defines behavioral risk assessment rules.
type BehavioralAssessmentRules struct {
	AccountAgeThresholds  []AccountAgeThreshold   `json:"account_age_thresholds"`
	ActivityMultipliers   map[string]float64      `json:"activity_multipliers"`
	PaymentHistoryWeight  float64                 `json:"payment_history_weight"`
}

// AccountAgeThreshold defines account age thresholds with risk multipliers.
type AccountAgeThreshold struct {
	MaxAge    time.Duration `json:"max_age"`
	Multiplier float64       `json:"multiplier"`
}

// FinancialRules defines financial risk assessment rules.
type FinancialRules struct {
	IncomeThresholds      []IncomeThreshold       `json:"income_thresholds"`
	CreditScoreThresholds []CreditScoreThreshold  `json:"credit_score_thresholds"`
	DebtToIncomeRatio     float64                 `json:"debt_to_income_ratio"`
}

// IncomeThreshold defines income thresholds with risk multipliers.
type IncomeThreshold struct {
	MinIncome  float64 `json:"min_income"`
	MaxIncome  float64 `json:"max_income"`
	Multiplier float64 `json:"multiplier"`
}

// CreditScoreThreshold defines credit score thresholds with risk multipliers.
type CreditScoreThreshold struct {
	MinScore   int     `json:"min_score"`
	MaxScore   int     `json:"max_score"`
	Multiplier float64 `json:"multiplier"`
}

// GeographicAssessmentRules defines geographic risk assessment rules.
type GeographicAssessmentRules struct {
	CountryRiskLevels     map[string]string       `json:"country_risk_levels"`
	RegionRiskLevels      map[string]string       `json:"region_risk_levels"`
	RiskLevelMultipliers  map[string]float64      `json:"risk_level_multipliers"`
}

// ProductRules defines product-specific risk assessment rules.
type ProductRules struct {
	CoverageThresholds    []CoverageThreshold     `json:"coverage_thresholds"`
	ProductTypeMultipliers map[string]float64     `json:"product_type_multipliers"`
	HighValueThreshold    float64                 `json:"high_value_threshold"`
}

// CoverageThreshold defines coverage amount thresholds with risk multipliers.
type CoverageThreshold struct {
	MinCoverage float64 `json:"min_coverage"`
	MaxCoverage float64 `json:"max_coverage"`
	Multiplier  float64 `json:"multiplier"`
}

// HistoricalRules defines historical risk assessment rules.
type HistoricalRules struct {
	ClaimHistoryWeight    float64                 `json:"claim_history_weight"`
	PaymentHistoryWeight  float64                 `json:"payment_history_weight"`
	PolicyHistoryWeight   float64                 `json:"policy_history_weight"`
	TimeDecayFactor       float64                 `json:"time_decay_factor"`
}

// LifestyleRules defines lifestyle risk assessment rules.
type LifestyleRules struct {
	OccupationRiskLevels  map[string]string       `json:"occupation_risk_levels"`
	HobbyRiskLevels       map[string]string       `json:"hobby_risk_levels"`
	RiskLevelMultipliers  map[string]float64      `json:"risk_level_multipliers"`
}

// ComplianceAssessmentRules defines compliance risk assessment rules.
type ComplianceAssessmentRules struct {
	KYCWeight             float64                 `json:"kyc_weight"`
	AMLWeight             float64                 `json:"aml_weight"`
	DataProtectionWeight  float64                 `json:"data_protection_weight"`
	StatusMultipliers     map[string]float64      `json:"status_multipliers"`
}

// ApprovalThresholds defines approval decision thresholds.
type ApprovalThresholds struct {
	AutoApproveMax        float64 `json:"auto_approve_max"`        // 40
	ConditionalMin        float64 `json:"conditional_min"`         // 40
	ConditionalMax        float64 `json:"conditional_max"`         // 60
	PendingReviewMin      float64 `json:"pending_review_min"`      // 60
	PendingReviewMax      float64 `json:"pending_review_max"`      // 80
	DeclineMin            float64 `json:"decline_min"`             // 80
}

// PremiumAdjustments defines premium adjustment rules.
type PremiumAdjustments struct {
	MaxIncrease           float64 `json:"max_increase"`            // 200%
	MaxDecrease           float64 `json:"max_decrease"`            // 50%
	BaseAdjustmentRate    float64 `json:"base_adjustment_rate"`    // 0.02 (2% per point)
	ImpactMultiplier      float64 `json:"impact_multiplier"`       // 1.0
}

// PricingConfig holds pricing engine configuration.
type PricingConfig struct {
	Enabled                bool                    `json:"enabled"`
	Version               string                  `json:"version"`
	BaseRates             map[string]float64      `json:"base_rates"`
	CoverageAdjustments   CoverageAdjustments     `json:"coverage_adjustments"`
	RiskAdjustments       RiskAdjustments         `json:"risk_adjustments"`
	DiscountRules         DiscountRules           `json:"discount_rules"`
	TaxRules              TaxRules                `json:"tax_rules"`
	FrequencyAdjustments  FrequencyAdjustments    `json:"frequency_adjustments"`
	MarketAdjustments     MarketAdjustments       `json:"market_adjustments"`
	LoyaltyAdjustments    LoyaltyAdjustments      `json:"loyalty_adjustments"`
	SeasonalAdjustments   SeasonalAdjustments     `json:"seasonal_adjustments"`
	ValidationRules       PricingValidationRules  `json:"validation_rules"`
}

// CoverageAdjustments defines coverage amount-based pricing adjustments.
type CoverageAdjustments struct {
	Thresholds            []CoverageThreshold     `json:"thresholds"`
	HighValueSurcharge    float64                 `json:"high_value_surcharge"`    // 0.001 (0.1%)
	VeryHighValueSurcharge float64                `json:"very_high_value_surcharge"` // 0.002 (0.2%)
}

// RiskAdjustments defines risk-based pricing adjustments.
type RiskAdjustments struct {
	NewCustomerSurcharge  float64                 `json:"new_customer_surcharge"`  // 0.02 (2%)
	EstablishedCustomerDiscount float64           `json:"established_customer_discount"` // 0.005 (0.5%)
	HighRiskSurcharge     float64                 `json:"high_risk_surcharge"`     // 0.05 (5%)
	LowRiskDiscount       float64                 `json:"low_risk_discount"`       // 0.03 (3%)
}

// DiscountRules defines discount rules and eligibility.
type DiscountRules struct {
	MultiPolicyDiscount   float64                 `json:"multi_policy_discount"`   // 0.10 (10%)
	LoyaltyDiscount       float64                 `json:"loyalty_discount"`        // 0.05 (5%)
	EarlyPaymentDiscount  float64                 `json:"early_payment_discount"`  // 0.03 (3%)
	SafeDriverDiscount    float64                 `json:"safe_driver_discount"`    // 0.08 (8%)
	SecuritySystemDiscount float64                `json:"security_system_discount"` // 0.06 (6%)
	BulkDiscount          float64                 `json:"bulk_discount"`           // 0.15 (15%)
}

// TaxRules defines tax calculation rules.
type TaxRules struct {
	DefaultRate           float64                 `json:"default_rate"`            // 0.08 (8%)
	JurisdictionRates     map[string]float64      `json:"jurisdiction_rates"`
	ProductTypeRates      map[string]float64      `json:"product_type_rates"`
	Exemptions            []string                `json:"exemptions"`
}

// FrequencyAdjustments defines payment frequency-based adjustments.
type FrequencyAdjustments struct {
	AnnualDiscount        float64                 `json:"annual_discount"`         // 0.05 (5%)
	QuarterlySurcharge    float64                 `json:"quarterly_surcharge"`     // 0.02 (2%)
	MonthlySurcharge      float64                 `json:"monthly_surcharge"`       // 0.05 (5%)
}

// MarketAdjustments defines market condition-based adjustments.
type MarketAdjustments struct {
	BaseAdjustment        float64                 `json:"base_adjustment"`         // 0.03 (3%)
	VolatilityMultiplier  float64                 `json:"volatility_multiplier"`   // 1.2
	EconomicIndicators    map[string]float64      `json:"economic_indicators"`
}

// LoyaltyAdjustments defines customer loyalty-based adjustments.
type LoyaltyAdjustments struct {
	LongTermDiscount      float64                 `json:"long_term_discount"`      // 0.08 (8%)
	MediumTermDiscount    float64                 `json:"medium_term_discount"`    // 0.03 (3%)
	AccountAgeThresholds  []AccountAgeThreshold   `json:"account_age_thresholds"`
}

// SeasonalAdjustments defines seasonal pricing adjustments.
type SeasonalAdjustments struct {
	WinterSurcharge       float64                 `json:"winter_surcharge"`        // 0.02 (2%)
	SummerSurcharge       float64                 `json:"summer_surcharge"`        // 0.01 (1%)
	SpringFallMultiplier  float64                 `json:"spring_fall_multiplier"`  // 1.0
}

// PricingValidationRules defines pricing validation rules.
type PricingValidationRules struct {
	MinPremium            float64                 `json:"min_premium"`             // 10.0
	MaxPremium            float64                 `json:"max_premium"`             // 1000000.0
	MaxAdjustment         float64                 `json:"max_adjustment"`          // 2.0 (200%)
	MinAdjustment         float64                 `json:"min_adjustment"`          // 0.1 (10%)
}

// UnderwritingConfig holds underwriting configuration.
type UnderwritingConfig struct {
	Enabled                bool                    `json:"enabled"`
	Version               string                  `json:"version"`
	DecisionThresholds    DecisionThresholds      `json:"decision_thresholds"`
	ConfidenceThresholds  ConfidenceThresholds    `json:"confidence_thresholds"`
	ConditionRules        ConditionRules          `json:"condition_rules"`
	ReviewRules           ReviewRules             `json:"review_rules"`
	ValidationRules       UnderwritingValidationRules `json:"validation_rules"`
}

// DecisionThresholds defines underwriting decision thresholds.
type DecisionThresholds struct {
	AutoApproveMax        float64 `json:"auto_approve_max"`        // 40
	ConditionalMin        float64 `json:"conditional_min"`         // 40
	ConditionalMax        float64 `json:"conditional_max"`         // 60
	PendingReviewMin      float64 `json:"pending_review_min"`      // 60
	PendingReviewMax      float64 `json:"pending_review_max"`      // 80
	DeclineMin            float64 `json:"decline_min"`             // 80
}

// ConditionRules defines conditional approval rules.
type ConditionRules struct {
	FinancialDocumentDays int     `json:"financial_document_days"` // 30
	MonitoringDays        int     `json:"monitoring_days"`         // 90
	InspectionDays        int     `json:"inspection_days"`         // 14
	PaymentDays           int     `json:"payment_days"`            // 7
	HighValueThreshold    float64 `json:"high_value_threshold"`    // 500000
}

// ReviewRules defines manual review rules.
type ReviewRules struct {
	SeniorReviewThreshold float64 `json:"senior_review_threshold"` // 100000
	ExecutiveReviewThreshold float64 `json:"executive_review_threshold"` // 500000
	ReviewTimeLimit       int     `json:"review_time_limit"`       // 5 days
}

// UnderwritingValidationRules defines underwriting validation rules.
type UnderwritingValidationRules struct {
	MinConfidence         float64 `json:"min_confidence"`          // 0.3
	MaxConfidence         float64 `json:"max_confidence"`          // 1.0
	MinRiskScore          float64 `json:"min_risk_score"`          // 0
	MaxRiskScore          float64 `json:"max_risk_score"`          // 100
	MinPremium            float64 `json:"min_premium"`             // 0
}

// CommissionConfig holds commission configuration.
type CommissionConfig struct {
	Enabled                bool                    `json:"enabled"`
	Version               string                  `json:"version"`
	DefaultRates          map[string]float64      `json:"default_rates"`
	MinAmounts            map[string]float64      `json:"min_amounts"`
	MaxAmounts            map[string]float64      `json:"max_amounts"`
	PaymentSchedules      PaymentSchedules        `json:"payment_schedules"`
	ValidationRules       CommissionValidationRules `json:"validation_rules"`
}

// PaymentSchedules defines commission payment schedules.
type PaymentSchedules struct {
	InitialDays           int     `json:"initial_days"`            // 30
	RenewalDays           int     `json:"renewal_days"`            // 45
	AdjustmentDays        int     `json:"adjustment_days"`         // 15
}

// CommissionValidationRules defines commission validation rules.
type CommissionValidationRules struct {
	MinRate               float64 `json:"min_rate"`                // 0
	MaxRate               float64 `json:"max_rate"`                // 100
	MinAmount             float64 `json:"min_amount"`              // 0
	MaxAmount             float64 `json:"max_amount"`              // 100000
}

// ComplianceConfig holds compliance configuration.
type ComplianceConfig struct {
	Enabled                bool                    `json:"enabled"`
	Version               string                  `json:"version"`
	KYCRequirements       KYCRequirements         `json:"kyc_requirements"`
	AMLRequirements       AMLRequirements         `json:"aml_requirements"`
	DataProtectionRules   DataProtectionRules     `json:"data_protection_rules"`
	RegulatoryRules       RegulatoryRules         `json:"regulatory_rules"`
	ValidationRules       ComplianceValidationRules `json:"validation_rules"`
}

// KYCRequirements defines KYC compliance requirements.
type KYCRequirements struct {
	RequiredFields        []string                `json:"required_fields"`
	DocumentTypes         []string                `json:"document_types"`
	VerificationMethods   []string                `json:"verification_methods"`
	UpdateFrequency       int                     `json:"update_frequency"`        // days
}

// AMLRequirements defines AML compliance requirements.
type AMLRequirements struct {
	MonitoringThresholds  []float64               `json:"monitoring_thresholds"`
	ReportingThresholds   []float64               `json:"reporting_thresholds"`
	ReviewPeriods         []int                   `json:"review_periods"`          // days
}

// DataProtectionRules defines data protection compliance rules.
type DataProtectionRules struct {
	RetentionPeriods      map[string]int          `json:"retention_periods"`       // days
	ConsentRequirements   []string                `json:"consent_requirements"`
	DataMinimizationRules map[string]int          `json:"data_minimization_rules"`
}

// RegulatoryRules defines regulatory compliance rules.
type RegulatoryRules struct {
	Jurisdictions         []string                `json:"jurisdictions"`
	ReportingRequirements []string                `json:"reporting_requirements"`
	AuditRequirements     []string                `json:"audit_requirements"`
}

// ComplianceValidationRules defines compliance validation rules.
type ComplianceValidationRules struct {
	MinScore              float64 `json:"min_score"`               // 0
	MaxScore              float64 `json:"max_score"`               // 100
	PassThreshold         float64 `json:"pass_threshold"`          // 90
	WarningThreshold      float64 `json:"warning_threshold"`       // 70
}

// PolicyLifecycleConfig holds policy lifecycle configuration.
type PolicyLifecycleConfig struct {
	Enabled                bool                    `json:"enabled"`
	Version               string                  `json:"version"`
	RenewalRules          RenewalRules            `json:"renewal_rules"`
	CancellationRules     CancellationRules       `json:"cancellation_rules"`
	GracePeriodRules      GracePeriodRules        `json:"grace_period_rules"`
	ValidationRules       PolicyLifecycleValidationRules `json:"validation_rules"`
}

// RenewalRules defines policy renewal rules.
type RenewalRules struct {
	AdvanceRenewalDays    int     `json:"advance_renewal_days"`    // 30
	RateIncreaseRate      float64 `json:"rate_increase_rate"`      // 0.03 (3%)
	FrequencyDiscounts    map[string]float64 `json:"frequency_discounts"`
	LoyaltyDiscounts      map[string]float64 `json:"loyalty_discounts"`
}

// CancellationRules defines policy cancellation rules.
type CancellationRules struct {
	CancellationFeeRate   float64 `json:"cancellation_fee_rate"`   // 0.10 (10%)
	RefundCalculationMethod string `json:"refund_calculation_method"` // pro_rated
	ProcessingDays        int     `json:"processing_days"`         // 7
}

// GracePeriodRules defines grace period rules.
type GracePeriodRules struct {
	DefaultDays           int     `json:"default_days"`            // 15
	PaymentFailureDays    int     `json:"payment_failure_days"`    // 30
	RenewalDays           int     `json:"renewal_days"`            // 15
}

// PolicyLifecycleValidationRules defines policy lifecycle validation rules.
type PolicyLifecycleValidationRules struct {
	MinEffectiveDate      int     `json:"min_effective_date"`      // 0 days
	MaxEffectiveDate      int     `json:"max_effective_date"`      // 365 days
	MinPolicyDuration     int     `json:"min_policy_duration"`     // 30 days
}

// ClaimProcessingConfig holds claim processing configuration.
type ClaimProcessingConfig struct {
	Enabled                bool                    `json:"enabled"`
	Version               string                  `json:"version"`
	WorkflowRules         WorkflowRules           `json:"workflow_rules"`
	ApprovalRules         ApprovalRules           `json:"approval_rules"`
	ValidationRules       ClaimProcessingValidationRules `json:"validation_rules"`
}

// WorkflowRules defines claim processing workflow rules.
type WorkflowRules struct {
	Stages                []WorkflowStage         `json:"stages"`
	ConditionalStages     []ConditionalStage      `json:"conditional_stages"`
	ParallelProcessing    bool                    `json:"parallel_processing"`
	TimeoutHours          int                     `json:"timeout_hours"`           // 72
}

// WorkflowStage defines a workflow stage configuration.
type WorkflowStage struct {
	ID                    string                  `json:"id"`
	Name                  string                  `json:"name"`
	Required              bool                    `json:"required"`
	AutoApproval          bool                    `json:"auto_approval"`
	TimeoutHours          int                     `json:"timeout_hours"`
	Conditions            map[string]interface{}  `json:"conditions"`
}

// ConditionalStage defines a conditional workflow stage.
type ConditionalStage struct {
	StageID               string                  `json:"stage_id"`
	Condition             string                  `json:"condition"`
	Threshold             float64                 `json:"threshold"`
	Field                 string                  `json:"field"`
}

// ApprovalRules defines claim approval rules.
type ApprovalRules struct {
	AutoApproveMax        float64 `json:"auto_approve_max"`        // 10000
	SeniorReviewThreshold float64 `json:"senior_review_threshold"` // 50000
	ExecutiveReviewThreshold float64 `json:"executive_review_threshold"` // 100000
	ManualReviewThreshold float64 `json:"manual_review_threshold"` // 250000
}

// ClaimProcessingValidationRules defines claim processing validation rules.
type ClaimProcessingValidationRules struct {
	MinClaimAmount        float64 `json:"min_claim_amount"`        // 0
	MaxClaimAmount        float64 `json:"max_claim_amount"`        // 10000000
	MaxReportingDelay     int     `json:"max_reporting_delay"`     // 365 days
	MinDocumentCount      int     `json:"min_document_count"`      // 1
}

