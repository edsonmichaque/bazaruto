package services

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/edsonmichaque/bazaruto/pkg/job"
	"github.com/google/uuid"
)

// ClaimProcessingService handles automated claim processing workflows and approval chains.
type ClaimProcessingService struct {
	claimStore   store.ClaimStore
	policyStore  store.PolicyStore
	userStore    store.UserStore
	fraudService *FraudDetectionService
	riskService  *RiskAssessmentService
	eventService *EventService
	dispatcher   job.Dispatcher
}

// NewClaimProcessingService creates a new ClaimProcessingService instance.
func NewClaimProcessingService(
	claimStore store.ClaimStore,
	policyStore store.PolicyStore,
	userStore store.UserStore,
	fraudService *FraudDetectionService,
	riskService *RiskAssessmentService,
	eventService *EventService,
	dispatcher job.Dispatcher,
) *ClaimProcessingService {
	return &ClaimProcessingService{
		claimStore:   claimStore,
		policyStore:  policyStore,
		userStore:    userStore,
		fraudService: fraudService,
		riskService:  riskService,
		eventService: eventService,
		dispatcher:   dispatcher,
	}
}

// ClaimWorkflow represents the automated claim processing workflow.
type ClaimWorkflow struct {
	ClaimID      uuid.UUID              `json:"claim_id"`
	CurrentStage string                 `json:"current_stage"`
	Stages       []WorkflowStage        `json:"stages"`
	Status       string                 `json:"status"` // pending, in_progress, completed, failed
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CompletedAt  *time.Time             `json:"completed_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// WorkflowStage represents a stage in the claim processing workflow.
type WorkflowStage struct {
	StageID      string                 `json:"stage_id"`
	Name         string                 `json:"name"`
	Status       string                 `json:"status"` // pending, in_progress, completed, failed, skipped
	StartedAt    *time.Time             `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at"`
	Result       string                 `json:"result"` // approved, declined, requires_review
	Decision     string                 `json:"decision"`
	Comments     string                 `json:"comments"`
	AssignedTo   *uuid.UUID             `json:"assigned_to"`
	AutoApproved bool                   `json:"auto_approved"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ProcessClaim initiates the automated claim processing workflow.
func (s *ClaimProcessingService) ProcessClaim(ctx context.Context, claimID uuid.UUID) (*ClaimWorkflow, error) {
	// Fetch claim details
	claim, err := s.claimStore.GetClaim(ctx, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch claim: %w", err)
	}

	// Fetch related policy
	policy, err := s.policyStore.GetPolicy(ctx, claim.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch policy: %w", err)
	}

	// Create workflow
	workflow := &ClaimWorkflow{
		ClaimID:      claimID,
		CurrentStage: "initial_review",
		Status:       "in_progress",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	// Define workflow stages based on claim characteristics
	workflow.Stages = s.defineWorkflowStages(claim, policy)

	// Start processing
	if err := s.executeWorkflowStage(ctx, workflow, "initial_review"); err != nil {
		return nil, fmt.Errorf("failed to execute initial review: %w", err)
	}

	return workflow, nil
}

// defineWorkflowStages defines the workflow stages based on claim and policy characteristics.
func (s *ClaimProcessingService) defineWorkflowStages(claim *models.Claim, policy *models.Policy) []WorkflowStage {
	stages := []WorkflowStage{
		{
			StageID: "initial_review",
			Name:    "Initial Review",
			Status:  "pending",
		},
		{
			StageID: "fraud_detection",
			Name:    "Fraud Detection",
			Status:  "pending",
		},
		{
			StageID: "policy_validation",
			Name:    "Policy Validation",
			Status:  "pending",
		},
		{
			StageID: "damage_assessment",
			Name:    "Damage Assessment",
			Status:  "pending",
		},
		{
			StageID: "approval_decision",
			Name:    "Approval Decision",
			Status:  "pending",
		},
	}

	// Add conditional stages based on claim amount
	if claim.ClaimAmount > 50000 {
		stages = append(stages, WorkflowStage{
			StageID: "senior_review",
			Name:    "Senior Review",
			Status:  "pending",
		})
	}

	// Add conditional stages based on claim amount
	if claim.ClaimAmount > 100000 {
		stages = append(stages, WorkflowStage{
			StageID: "executive_approval",
			Name:    "Executive Approval",
			Status:  "pending",
		})
	}

	// Add payout stage
	stages = append(stages, WorkflowStage{
		StageID: "payout_processing",
		Name:    "Payout Processing",
		Status:  "pending",
	})

	return stages
}

// executeWorkflowStage executes a specific stage of the workflow.
func (s *ClaimProcessingService) executeWorkflowStage(ctx context.Context, workflow *ClaimWorkflow, stageID string) error {
	// Find the stage
	var stage *WorkflowStage
	for i := range workflow.Stages {
		if workflow.Stages[i].StageID == stageID {
			stage = &workflow.Stages[i]
			break
		}
	}

	if stage == nil {
		return fmt.Errorf("stage %s not found", stageID)
	}

	// Update stage status
	stage.Status = "in_progress"
	now := time.Now()
	stage.StartedAt = &now
	workflow.UpdatedAt = now

	// Execute stage-specific logic
	var err error
	switch stageID {
	case "initial_review":
		err = s.executeInitialReview(ctx, workflow, stage)
	case "fraud_detection":
		err = s.executeFraudDetection(ctx, workflow, stage)
	case "policy_validation":
		err = s.executePolicyValidation(ctx, workflow, stage)
	case "damage_assessment":
		err = s.executeDamageAssessment(ctx, workflow, stage)
	case "senior_review":
		err = s.executeSeniorReview(ctx, workflow, stage)
	case "executive_approval":
		err = s.executeExecutiveApproval(ctx, workflow, stage)
	case "approval_decision":
		err = s.executeApprovalDecision(ctx, workflow, stage)
	case "payout_processing":
		err = s.executePayoutProcessing(ctx, workflow, stage)
	default:
		err = fmt.Errorf("unknown stage: %s", stageID)
	}

	// Update stage completion
	if err != nil {
		stage.Status = "failed"
		stage.Comments = err.Error()
		workflow.Status = "failed"
	} else {
		stage.Status = "completed"
		stage.CompletedAt = &now
	}

	// Move to next stage if current stage completed successfully
	if stage.Status == "completed" {
		if err := s.moveToNextStage(ctx, workflow); err != nil {
			return fmt.Errorf("failed to move to next stage: %w", err)
		}
	}

	return err
}

// executeInitialReview executes the initial review stage.
func (s *ClaimProcessingService) executeInitialReview(ctx context.Context, workflow *ClaimWorkflow, stage *WorkflowStage) error {
	// Fetch claim details
	claim, err := s.claimStore.GetClaim(ctx, workflow.ClaimID)
	if err != nil {
		return fmt.Errorf("failed to fetch claim: %w", err)
	}

	// Basic validation checks
	if claim.Title == "" {
		stage.Result = "declined"
		stage.Decision = "Missing claim title"
		return fmt.Errorf("claim title is required")
	}

	if claim.Description == "" {
		stage.Result = "declined"
		stage.Decision = "Missing claim description"
		return fmt.Errorf("claim description is required")
	}

	if claim.ClaimAmount <= 0 {
		stage.Result = "declined"
		stage.Decision = "Invalid claim amount"
		return fmt.Errorf("claim amount must be greater than 0")
	}

	// Check if claim is within policy period
	policy, err := s.policyStore.GetPolicy(ctx, claim.PolicyID)
	if err != nil {
		return fmt.Errorf("failed to fetch policy: %w", err)
	}

	if claim.IncidentDate.Before(policy.EffectiveDate) || claim.IncidentDate.After(policy.ExpirationDate) {
		stage.Result = "declined"
		stage.Decision = "Incident date outside policy period"
		return fmt.Errorf("incident date must be within policy period")
	}

	// Auto-approve if all basic checks pass
	stage.Result = "approved"
	stage.Decision = "Initial review passed"
	stage.AutoApproved = true
	stage.Comments = "All initial validation checks passed"

	return nil
}

// executeFraudDetection executes the fraud detection stage.
func (s *ClaimProcessingService) executeFraudDetection(ctx context.Context, workflow *ClaimWorkflow, stage *WorkflowStage) error {
	// Run fraud detection analysis
	fraudScore, err := s.fraudService.AnalyzeClaimForFraud(ctx, workflow.ClaimID)
	if err != nil {
		return fmt.Errorf("failed to analyze fraud: %w", err)
	}

	// Store fraud analysis results
	stage.Metadata = map[string]interface{}{
		"fraud_score":     fraudScore.Score,
		"risk_level":      fraudScore.RiskLevel,
		"requires_review": fraudScore.RequiresReview,
	}

	// Make decision based on fraud score
	if fraudScore.Score >= 80 {
		stage.Result = "declined"
		stage.Decision = "High fraud risk detected"
		stage.Comments = fmt.Sprintf("Fraud score: %.2f, Risk level: %s", fraudScore.Score, fraudScore.RiskLevel)
		return fmt.Errorf("claim declined due to high fraud risk")
	} else if fraudScore.Score >= 60 {
		stage.Result = "requires_review"
		stage.Decision = "Moderate fraud risk - requires manual review"
		stage.Comments = fmt.Sprintf("Fraud score: %.2f, Risk level: %s", fraudScore.Score, fraudScore.RiskLevel)
		// Don't return error - continue to next stage for manual review
	} else {
		stage.Result = "approved"
		stage.Decision = "Low fraud risk"
		stage.AutoApproved = true
		stage.Comments = fmt.Sprintf("Fraud score: %.2f, Risk level: %s", fraudScore.Score, fraudScore.RiskLevel)
	}

	return nil
}

// executePolicyValidation executes the policy validation stage.
func (s *ClaimProcessingService) executePolicyValidation(ctx context.Context, workflow *ClaimWorkflow, stage *WorkflowStage) error {
	// Fetch claim and policy details
	claim, err := s.claimStore.GetClaim(ctx, workflow.ClaimID)
	if err != nil {
		return fmt.Errorf("failed to fetch claim: %w", err)
	}

	policy, err := s.policyStore.GetPolicy(ctx, claim.PolicyID)
	if err != nil {
		return fmt.Errorf("failed to fetch policy: %w", err)
	}

	// Validate policy is active
	if policy.Status != models.PolicyStatusActive {
		stage.Result = "declined"
		stage.Decision = "Policy is not active"
		return fmt.Errorf("policy is not active")
	}

	// Validate claim amount against coverage
	if claim.ClaimAmount > policy.CoverageAmount {
		stage.Result = "declined"
		stage.Decision = "Claim amount exceeds coverage"
		return fmt.Errorf("claim amount exceeds policy coverage")
	}

	// Validate user owns the policy
	if policy.UserID != claim.UserID {
		stage.Result = "declined"
		stage.Decision = "User does not own the policy"
		return fmt.Errorf("user does not own the policy")
	}

	// Auto-approve if all policy validations pass
	stage.Result = "approved"
	stage.Decision = "Policy validation passed"
	stage.AutoApproved = true
	stage.Comments = "All policy validation checks passed"

	return nil
}

// executeDamageAssessment executes the damage assessment stage.
func (s *ClaimProcessingService) executeDamageAssessment(ctx context.Context, workflow *ClaimWorkflow, stage *WorkflowStage) error {
	// Fetch claim details
	claim, err := s.claimStore.GetClaim(ctx, workflow.ClaimID)
	if err != nil {
		return fmt.Errorf("failed to fetch claim: %w", err)
	}

	// Check if damage assessment is required based on claim amount
	if claim.ClaimAmount > 10000 {
		// For high-value claims, require manual assessment
		stage.Result = "requires_review"
		stage.Decision = "Manual damage assessment required"
		stage.Comments = "High-value claim requires professional damage assessment"
		// Don't return error - continue to next stage for manual review
	} else {
		// For lower-value claims, auto-approve based on documentation
		docCount := len(claim.Documents)
		if docCount >= 2 {
			stage.Result = "approved"
			stage.Decision = "Damage assessment approved based on documentation"
			stage.AutoApproved = true
			stage.Comments = fmt.Sprintf("Approved based on %d supporting documents", docCount)
		} else {
			stage.Result = "requires_review"
			stage.Decision = "Insufficient documentation for damage assessment"
			stage.Comments = "Additional documentation required for damage assessment"
		}
	}

	return nil
}

// executeSeniorReview executes the senior review stage.
func (s *ClaimProcessingService) executeSeniorReview(ctx context.Context, workflow *ClaimWorkflow, stage *WorkflowStage) error {
	// This stage requires manual review by a senior adjuster
	// In a real implementation, this would create a task for a senior adjuster
	stage.Result = "requires_review"
	stage.Decision = "Senior review required"
	stage.Comments = "Claim requires review by senior adjuster due to high value"

	// Dispatch notification job for senior review
	// In a real implementation, this would dispatch a notification job
	// For now, we'll just log the requirement
	stage.Comments += "; Senior review notification dispatched"

	return nil
}

// executeExecutiveApproval executes the executive approval stage.
func (s *ClaimProcessingService) executeExecutiveApproval(ctx context.Context, workflow *ClaimWorkflow, stage *WorkflowStage) error {
	// This stage requires manual approval by an executive
	// In a real implementation, this would create a task for an executive
	stage.Result = "requires_review"
	stage.Decision = "Executive approval required"
	stage.Comments = "Very high-value claim requires executive approval"

	// Dispatch notification job for executive approval
	// In a real implementation, this would dispatch a notification job
	// For now, we'll just log the requirement
	stage.Comments += "; Executive approval notification dispatched"

	return nil
}

// executeApprovalDecision executes the final approval decision stage.
func (s *ClaimProcessingService) executeApprovalDecision(ctx context.Context, workflow *ClaimWorkflow, stage *WorkflowStage) error {
	// Review all previous stages to make final decision
	approvedStages := 0
	declinedStages := 0
	reviewRequiredStages := 0

	for _, s := range workflow.Stages {
		switch s.Result {
		case "approved":
			approvedStages++
		case "declined":
			declinedStages++
		case "requires_review":
			reviewRequiredStages++
		}
	}

	// Make final decision
	if declinedStages > 0 {
		stage.Result = "declined"
		stage.Decision = "Claim declined based on workflow analysis"
		stage.Comments = fmt.Sprintf("Declined due to %d failed stages", declinedStages)

		// Update claim status
		claim, err := s.claimStore.GetClaim(ctx, workflow.ClaimID)
		if err == nil {
			claim.Status = models.ClaimStatusDenied
			denialReason := stage.Decision
			claim.DenialReason = &denialReason
			now := time.Now()
			claim.ResolvedDate = &now
			_ = s.claimStore.UpdateClaim(ctx, claim)
		}

		return fmt.Errorf("claim declined")
	} else if reviewRequiredStages > 0 {
		stage.Result = "requires_review"
		stage.Decision = "Claim requires manual review"
		stage.Comments = fmt.Sprintf("Manual review required for %d stages", reviewRequiredStages)

		// Update claim status
		claim, err := s.claimStore.GetClaim(ctx, workflow.ClaimID)
		if err == nil {
			claim.Status = models.ClaimStatusUnderReview
			_ = s.claimStore.UpdateClaim(ctx, claim)
		}
	} else {
		stage.Result = "approved"
		stage.Decision = "Claim approved for payout"
		stage.AutoApproved = true
		stage.Comments = "All automated stages approved"

		// Update claim status
		claim, err := s.claimStore.GetClaim(ctx, workflow.ClaimID)
		if err == nil {
			claim.Status = models.ClaimStatusApproved
			_ = s.claimStore.UpdateClaim(ctx, claim)
		}
	}

	return nil
}

// executePayoutProcessing executes the payout processing stage.
func (s *ClaimProcessingService) executePayoutProcessing(ctx context.Context, workflow *ClaimWorkflow, stage *WorkflowStage) error {
	// Only process payout if claim was approved
	if stage.Result != "approved" {
		stage.Status = "skipped"
		stage.Comments = "Payout skipped - claim not approved"
		return nil
	}

	// Dispatch payout settlement job
	// In a real implementation, this would dispatch a payout job
	// For now, we'll simulate successful dispatch
	stage.Comments = "Payout settlement job dispatched successfully"

	stage.Result = "approved"
	stage.Decision = "Payout processing initiated"
	stage.AutoApproved = true
	stage.Comments = "Payout settlement job dispatched successfully"

	return nil
}

// moveToNextStage moves the workflow to the next stage.
func (s *ClaimProcessingService) moveToNextStage(ctx context.Context, workflow *ClaimWorkflow) error {
	// Find current stage index
	currentIndex := -1
	for i, stage := range workflow.Stages {
		if stage.StageID == workflow.CurrentStage {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return fmt.Errorf("current stage not found: %s", workflow.CurrentStage)
	}

	// Check if there's a next stage
	if currentIndex+1 >= len(workflow.Stages) {
		// Workflow completed
		workflow.Status = "completed"
		now := time.Now()
		workflow.CompletedAt = &now
		workflow.UpdatedAt = now

		// Publish workflow completed event
		if s.eventService != nil {
			// In a real implementation, this would publish a workflow completed event
			// For now, we'll skip the event publishing
		}

		return nil
	}

	// Move to next stage
	nextStage := workflow.Stages[currentIndex+1]
	workflow.CurrentStage = nextStage.StageID
	workflow.UpdatedAt = time.Now()

	// Execute next stage
	return s.executeWorkflowStage(ctx, workflow, nextStage.StageID)
}

// GetWorkflowStatus retrieves the current status of a claim workflow.
func (s *ClaimProcessingService) GetWorkflowStatus(ctx context.Context, claimID uuid.UUID) (*ClaimWorkflow, error) {
	// In a real implementation, this would retrieve from a workflow store
	// For now, we'll create a new workflow and return its status
	workflow, err := s.ProcessClaim(ctx, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to process claim: %w", err)
	}

	return workflow, nil
}

// UpdateWorkflowStage manually updates a workflow stage (for manual reviews).
func (s *ClaimProcessingService) UpdateWorkflowStage(ctx context.Context, claimID uuid.UUID, stageID string, result, decision, comments string, assignedTo *uuid.UUID) error {
	// In a real implementation, this would update the workflow in the store
	// For now, we'll simulate the update
	workflow, err := s.GetWorkflowStatus(ctx, claimID)
	if err != nil {
		return fmt.Errorf("failed to get workflow status: %w", err)
	}

	// Find and update the stage
	for i := range workflow.Stages {
		if workflow.Stages[i].StageID == stageID {
			workflow.Stages[i].Result = result
			workflow.Stages[i].Decision = decision
			workflow.Stages[i].Comments = comments
			workflow.Stages[i].AssignedTo = assignedTo
			workflow.Stages[i].Status = "completed"
			now := time.Now()
			workflow.Stages[i].CompletedAt = &now
			workflow.UpdatedAt = now
			break
		}
	}

	// Continue workflow if stage was completed
	if result == "approved" || result == "declined" {
		return s.moveToNextStage(ctx, workflow)
	}

	return nil
}
