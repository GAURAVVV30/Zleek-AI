package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/hcl-backend/services/api-go/internal/projects/application"
	"github.com/hcl-backend/services/api-go/internal/projects/domain"
)

type mockAssessmentSvc struct{}

func (m *mockAssessmentSvc) GetProjectDefinition(ctx context.Context, conceptID string) (*domain.Project, error) {
	if conceptID == "invalid" {
		return nil, domain.ErrNoProjectForConcept
	}
	return &domain.Project{
		ID:        "p-1",
		ConceptID: conceptID,
	}, nil
}

type mockEvidenceSvc struct{}

func (m *mockEvidenceSvc) RecordProjectSubmission(ctx context.Context, submission *domain.ProjectSubmission) error {
	return nil
}

func (m *mockEvidenceSvc) GetProjectStatus(ctx context.Context, learnerID, conceptID string) (*domain.ProjectState, error) {
	if conceptID == "none" {
		return nil, nil
	}
	score := 85.0
	return &domain.ProjectState{
		Status:    domain.ProjectStatusReviewed,
		Result:    "competent",
		Score:     &score,
		UpdatedAt: time.Now(),
	}, nil
}

type mockStorageSvc struct{}

func (m *mockStorageSvc) ValidateArtifactReference(ctx context.Context, reference string) error {
	if reference == "invalid" {
		return domain.ErrInvalidArtifact
	}
	return nil
}

func TestGetProject(t *testing.T) {
	uc := application.NewGetProjectUseCase(&mockAssessmentSvc{})
	proj, err := uc.Execute(context.Background(), "c-1")
	if err != nil {
		t.Fatal(err)
	}
	if proj.ID != "p-1" {
		t.Fatal("wrong project ID")
	}

	_, err = uc.Execute(context.Background(), "invalid")
	if err != domain.ErrNoProjectForConcept {
		t.Fatal("expected err")
	}
}

func TestSubmitProject(t *testing.T) {
	uc := application.NewSubmitProjectUseCase(
		&mockAssessmentSvc{},
		&mockStorageSvc{},
		&mockEvidenceSvc{},
	)

	err := uc.Execute(context.Background(), "l-1", "c-1", domain.SubmissionMetadata{ArtifactReference: "s3://foo"})
	if err != nil {
		t.Fatal(err)
	}

	err = uc.Execute(context.Background(), "l-1", "invalid", domain.SubmissionMetadata{ArtifactReference: "s3://foo"})
	if err != domain.ErrNoProjectForConcept {
		t.Fatal("expected err for invalid concept")
	}

	err = uc.Execute(context.Background(), "l-1", "c-1", domain.SubmissionMetadata{ArtifactReference: "invalid"})
	if err != domain.ErrInvalidArtifact {
		t.Fatal("expected err for invalid artifact")
	}
}

func TestGetProjectStatus(t *testing.T) {
	uc := application.NewGetProjectStatusUseCase(&mockEvidenceSvc{})
	state, err := uc.Execute(context.Background(), "l-1", "c-1")
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != domain.ProjectStatusReviewed {
		t.Fatal("wrong state")
	}
}
