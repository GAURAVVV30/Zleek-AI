package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/projects/domain"
)

type AssessmentService interface {
	GetProjectDefinition(ctx context.Context, conceptID string) (*domain.Project, error)
}

type EvidenceService interface {
	RecordProjectSubmission(ctx context.Context, submission *domain.ProjectSubmission) error
	GetProjectStatus(ctx context.Context, learnerID, conceptID string) (*domain.ProjectState, error)
}

type StorageService interface {
	ValidateArtifactReference(ctx context.Context, reference string) error
}
