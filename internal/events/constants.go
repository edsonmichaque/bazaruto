package events

import "time"

// Default configuration constants
const (
	// Event processing
	DefaultMaxEvents = 10000
	DefaultEventTTL  = 24 * time.Hour

	// Event types
	EventTypeUserRegistered = "user.registered"
	EventTypeUserLoggedIn   = "user.logged_in"
	EventTypeUserUpdated    = "user.updated"
	EventTypeUserDeleted    = "user.deleted"

	EventTypeQuoteCreated    = "quote.created"
	EventTypeQuoteCalculated = "quote.calculated"
	EventTypeQuoteExpired    = "quote.expired"
	EventTypeQuoteAccepted   = "quote.accepted"
	EventTypeQuoteRejected   = "quote.rejected"

	EventTypePaymentInitiated = "payment.initiated"
	EventTypePaymentCompleted = "payment.completed"
	EventTypePaymentFailed    = "payment.failed"
	EventTypePaymentRefunded  = "payment.refunded"

	EventTypePolicyCreated   = "policy.created"
	EventTypePolicyActivated = "policy.activated"
	EventTypePolicyExpired   = "policy.expired"
	EventTypePolicyCancelled = "policy.cancelled"
	EventTypePolicyRenewed   = "policy.renewed"

	EventTypeClaimSubmitted = "claim.submitted"
	EventTypeClaimApproved  = "claim.approved"
	EventTypeClaimRejected  = "claim.rejected"
	EventTypeClaimSettled   = "claim.settled"
	EventTypeClaimClosed    = "claim.closed"

	EventTypeFraudDetected = "fraud.detected"
	EventTypeFraudAnalysis = "fraud.analysis"
	EventTypeRiskAssessed  = "risk.assessed"
	EventTypeUnderwriting  = "underwriting.completed"
	EventTypeCommission    = "commission.calculated"
	EventTypeCompliance    = "compliance.checked"

	// Entity types
	EntityTypeUser       = "user"
	EntityTypeQuote      = "quote"
	EntityTypePayment    = "payment"
	EntityTypePolicy     = "policy"
	EntityTypeClaim      = "claim"
	EntityTypeFraud      = "fraud"
	EntityTypeRisk       = "risk"
	EntityTypeCommission = "commission"

	// Event versions
	EventVersionV1 = "1.0"
	EventVersionV2 = "2.0"

	// Event metadata keys
	MetadataKeySource        = "source"
	MetadataKeyUserAgent     = "user_agent"
	MetadataKeyIPAddress     = "ip_address"
	MetadataKeySessionID     = "session_id"
	MetadataKeyRequestID     = "request_id"
	MetadataKeyTraceID       = "trace_id"
	MetadataKeySpanID        = "span_id"
	MetadataKeyCorrelationID = "correlation_id"

	// Event processing
	EventProcessingTimeout = 30 * time.Second
	EventRetryAttempts     = 3
	EventRetryDelay        = 1 * time.Second

	// Event store
	EventStoreBatchSize     = 100
	EventStoreFlushInterval = 5 * time.Second
	EventStoreMaxRetries    = 5

	// Event bus
	EventBusMaxConcurrency = 100
	EventBusQueueSize      = 1000
	EventBusWorkerTimeout  = 10 * time.Second

	// Event handlers
	EventHandlerTimeout    = 30 * time.Second
	EventHandlerMaxRetries = 3
	EventHandlerRetryDelay = 2 * time.Second

	// Event monitoring
	EventMetricsInterval     = 1 * time.Minute
	EventHealthCheckInterval = 30 * time.Second

	// Event cleanup
	EventCleanupInterval = 1 * time.Hour
	EventRetentionPeriod = 30 * 24 * time.Hour // 30 days
	EventArchivePeriod   = 90 * 24 * time.Hour // 90 days

	// Event validation
	EventMaxPayloadSize       = 1024 * 1024 // 1MB
	EventMaxMetadataSize      = 64 * 1024   // 64KB
	EventMaxDescriptionLength = 1000

	// Event security
	EventEncryptionKeySize = 32
	EventSigningKeySize    = 64
	EventTokenExpiry       = 1 * time.Hour

	// Event notifications
	EventNotificationTimeout = 10 * time.Second
	EventNotificationRetries = 3
	EventNotificationDelay   = 1 * time.Second

	// Event audit
	EventAuditRetentionPeriod = 365 * 24 * time.Hour // 1 year
	EventAuditBatchSize       = 50
	EventAuditFlushInterval   = 1 * time.Minute

	// Event debugging
	EventDebugMode        = false
	EventDebugLogLevel    = "debug"
	EventDebugMaxLogSize  = 10 * 1024 * 1024 // 10MB
	EventDebugLogRotation = 5

	// Event performance
	EventPerformanceThreshold = 100 * time.Millisecond
	EventSlowQueryThreshold   = 1 * time.Second
	EventMemoryThreshold      = 100 * 1024 * 1024 // 100MB

	// Event reliability
	EventReliabilityTarget = 99.9 // 99.9% uptime
	EventErrorThreshold    = 0.1  // 0.1% error rate
	EventLatencyThreshold  = 500 * time.Millisecond

	// Event scaling
	EventMinWorkers           = 1
	EventMaxWorkers           = 100
	EventWorkerScaleUpDelay   = 30 * time.Second
	EventWorkerScaleDownDelay = 5 * time.Minute

	// Event partitioning
	EventPartitionCount            = 10
	EventPartitionKeySize          = 32
	EventPartitionBalanceThreshold = 0.1 // 10% imbalance tolerance

	// Event serialization
	EventSerializationFormat = "json"
	EventCompressionEnabled  = true
	EventCompressionLevel    = 6

	// Event routing
	EventRoutingTimeout                 = 5 * time.Second
	EventRoutingRetries                 = 3
	EventRoutingCircuitBreakerThreshold = 5

	// Event aggregation
	EventAggregationWindow        = 1 * time.Minute
	EventAggregationBatchSize     = 1000
	EventAggregationFlushInterval = 10 * time.Second

	// Event correlation
	EventCorrelationWindow    = 5 * time.Minute
	EventCorrelationMaxEvents = 1000
	EventCorrelationTimeout   = 30 * time.Second

	// Event enrichment
	EventEnrichmentTimeout  = 5 * time.Second
	EventEnrichmentRetries  = 2
	EventEnrichmentCacheTTL = 1 * time.Hour

	// Event transformation
	EventTransformationTimeout   = 10 * time.Second
	EventTransformationRetries   = 3
	EventTransformationCacheSize = 1000

	// Event validation rules
	EventValidationStrict        = true
	EventValidationSchemaVersion = "1.0"
	EventValidationCustomRules   = false

	// Event error handling
	EventErrorRecoveryEnabled = true
	EventErrorRecoveryTimeout = 1 * time.Minute
	EventErrorRecoveryRetries = 5

	// Event monitoring and alerting
	EventAlertThreshold    = 10 // errors per minute
	EventAlertCooldown     = 5 * time.Minute
	EventAlertMaxFrequency = 1 * time.Hour

	// Event testing
	EventTestMode           = false
	EventTestTimeout        = 30 * time.Second
	EventTestRetries        = 1
	EventTestCleanupEnabled = true
)
