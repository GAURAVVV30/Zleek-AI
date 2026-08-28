package application_test

import (
	"context"
	"testing"

	"github.com/hcl-backend/services/api-go/internal/resources/application"
	"github.com/hcl-backend/services/api-go/internal/resources/domain"
)

type mockResourceRepo struct {
	resources map[string]*domain.Resource
}

func newMockResourceRepo() *mockResourceRepo {
	return &mockResourceRepo{
		resources: make(map[string]*domain.Resource),
	}
}

func (m *mockResourceRepo) GetResource(ctx context.Context, id string) (*domain.Resource, error) {
	if r, ok := m.resources[id]; ok {
		return r, nil
	}
	return nil, domain.ErrResourceNotFound
}

func (m *mockResourceRepo) CreateResource(ctx context.Context, resource *domain.Resource, concepts []domain.ResourceConcept) error {
	m.resources[resource.ID] = resource
	return nil
}

func (m *mockResourceRepo) UpdateResource(ctx context.Context, resource *domain.Resource) error {
	if _, ok := m.resources[resource.ID]; !ok {
		return domain.ErrResourceNotFound
	}
	m.resources[resource.ID] = resource
	return nil
}

func (m *mockResourceRepo) ListResources(ctx context.Context) ([]domain.Resource, error) {
	var list []domain.Resource
	for _, r := range m.resources {
		list = append(list, *r)
	}
	return list, nil
}

func (m *mockResourceRepo) GetFeedbackSignals(ctx context.Context, resourceID string) (*domain.ResourceQualitySignal, error) {
	return &domain.ResourceQualitySignal{ResourceID: resourceID}, nil
}

type mockConceptValidator struct{}

func (m *mockConceptValidator) ValidateConcepts(ctx context.Context, conceptIDs []string) error {
	return nil
}

func TestCreateResource(t *testing.T) {
	repo := newMockResourceRepo()
	validator := &mockConceptValidator{}
	uc := application.NewCreateResourceUseCase(repo, validator)

	cmd := application.CreateResourceCommand{
		URL:          "https://example.com/test",
		ResourceType: "article",
	}

	res, err := uc.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Status != domain.StatusCandidate {
		t.Fatalf("expected candidate status, got %v", res.Status)
	}
}

func TestUpdateResourcePublish(t *testing.T) {
	repo := newMockResourceRepo()
	repo.resources["r-1"] = &domain.Resource{
		ID:     "r-1",
		Status: domain.StatusCandidate,
	}

	uc := application.NewUpdateResourceUseCase(repo)

	statusPub := string(domain.StatusPublished)
	curatorID := "c-1"

	cmd := application.UpdateResourceCommand{
		ID:        "r-1",
		Status:    &statusPub,
		CuratorID: &curatorID,
	}

	res, err := uc.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.Status != domain.StatusPublished {
		t.Fatalf("expected published, got %v", res.Status)
	}
	if res.CuratedBy == nil || *res.CuratedBy != curatorID {
		t.Fatalf("expected curated_by to be set")
	}
}
