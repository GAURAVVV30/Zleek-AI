package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/hcl-backend/services/api-go/internal/progress/domain"
)

type GetCompletionBadgeUseCase struct {
	repo  ProgressRepository
	goals GoalService
}

func NewGetCompletionBadgeUseCase(repo ProgressRepository, goals GoalService) *GetCompletionBadgeUseCase {
	return &GetCompletionBadgeUseCase{repo: repo, goals: goals}
}

func (uc *GetCompletionBadgeUseCase) Execute(ctx context.Context, learnerID string) (*domain.CompletionBadgeResponse, error) {
	_, goalTitle, structureID, err := uc.goals.ActiveStructureMeta(ctx, learnerID)
	if err != nil || structureID == "" {
		return &domain.CompletionBadgeResponse{
			Eligible:         false,
			Role:             "Learner Track",
			CompletedModules: 0,
			TotalModules:     0,
			Message:          "No active goal or roadmap found for the authenticated learner.",
		}, nil
	}

	totalModules, completedModules, err := uc.repo.GetCompletionBadgeStatus(ctx, learnerID, structureID)
	if err != nil {
		return nil, err
	}

	roleName := cleanRoleTitle(goalTitle)
	eligible := (totalModules > 0 && completedModules == totalModules)

	if !eligible {
		return &domain.CompletionBadgeResponse{
			Eligible:         false,
			Role:             roleName,
			CompletedModules: completedModules,
			TotalModules:     totalModules,
			Message:          fmt.Sprintf("All %d roadmap modules must be completed to unlock the Role Completion Badge.", totalModules),
		}, nil
	}

	return &domain.CompletionBadgeResponse{
		Eligible:         true,
		Role:             roleName,
		CompletedModules: completedModules,
		TotalModules:     totalModules,
		Badge: &domain.BadgeDetails{
			Title:  roleName,
			Status: "verified",
		},
	}, nil
}

func cleanRoleTitle(goalTitle string) string {
	raw := strings.TrimSpace(goalTitle)
	if raw == "" {
		return "LEARNER"
	}
	lower := strings.ToLower(raw)

	if strings.Contains(lower, "ai engineer") || strings.Contains(lower, "ai_engineer") {
		return "AI ENGINEER"
	}
	if strings.Contains(lower, "data scientist") || strings.Contains(lower, "ai_data_scientist") {
		return "AI DATA SCIENTIST"
	}
	if strings.Contains(lower, "data analyst") || strings.Contains(lower, "data_analyst") {
		return "DATA ANALYST"
	}
	if strings.Contains(lower, "backend engineer") || strings.Contains(lower, "backend_engineer") {
		return "BACKEND ENGINEER"
	}
	if strings.Contains(lower, "frontend engineer") || strings.Contains(lower, "frontend_engineer") {
		return "FRONTEND ENGINEER"
	}
	if strings.Contains(lower, "full stack") || strings.Contains(lower, "full_stack") {
		return "FULL STACK ENGINEER"
	}
	if strings.Contains(lower, "devops") || strings.Contains(lower, "devops_sre") {
		return "DEVOPS & SRE"
	}
	if strings.Contains(lower, "machine learning") || strings.Contains(lower, "machine_learning") {
		return "MACHINE LEARNING ENGINEER"
	}
	if strings.Contains(lower, "mobile engineer") || strings.Contains(lower, "mobile_engineer") {
		return "MOBILE ENGINEER"
	}
	if strings.Contains(lower, "product manager") || strings.Contains(lower, "product_manager") {
		return "PRODUCT MANAGER"
	}
	if strings.Contains(lower, "software architect") || strings.Contains(lower, "software_architect") {
		return "SOFTWARE ARCHITECT"
	}

	// Extract title from "I want to master X" or "Become an? X"
	cleaned := strings.TrimPrefix(raw, "I want to master ")
	cleaned = strings.TrimPrefix(cleaned, "Become an ")
	cleaned = strings.TrimPrefix(cleaned, "Become a ")
	cleaned = strings.TrimSuffix(cleaned, " from the ground up.")
	cleaned = strings.TrimSuffix(cleaned, ".")

	return strings.ToUpper(strings.TrimSpace(cleaned))
}
