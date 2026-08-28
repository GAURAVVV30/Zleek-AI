package application_test

import (
	"context"
	"testing"

	"github.com/hcl-backend/services/api-go/internal/admin/application"
	"github.com/hcl-backend/services/api-go/internal/admin/domain"
)

type mockIdentityService struct {
	users map[string]*domain.User
}

func (m *mockIdentityService) ListUsers(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	for _, u := range m.users {
		users = append(users, *u)
	}
	return users, nil
}

func (m *mockIdentityService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockIdentityService) UpdateUser(ctx context.Context, id string, role, status *string) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	if role != nil {
		u.Role = *role
	}
	if status != nil {
		u.Status = *status
	}
	return u, nil
}

type mockAuditRepo struct {
	records []domain.AuditRecord
}

func (m *mockAuditRepo) Create(ctx context.Context, record *domain.AuditRecord) error {
	record.ID = "audit-1"
	m.records = append(m.records, *record)
	return nil
}

func (m *mockAuditRepo) List(ctx context.Context) ([]domain.AuditRecord, error) {
	return m.records, nil
}

func TestUpdateUser(t *testing.T) {
	idSvc := &mockIdentityService{
		users: map[string]*domain.User{
			"u1": {ID: "u1", Role: "learner", Status: "active"},
			"a1": {ID: "a1", Role: "admin", Status: "active"},
		},
	}
	auditRepo := &mockAuditRepo{}

	uc := application.NewUpdateUserUseCase(idSvc, auditRepo)

	// Valid update
	role := "curator"
	updated, err := uc.Execute(context.Background(), "a1", "u1", &role, nil)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Role != "curator" {
		t.Fatal("expected curator")
	}
	if len(auditRepo.records) != 1 {
		t.Fatal("expected 1 audit record")
	}

	// Invalid role
	invalidRole := "foo"
	_, err = uc.Execute(context.Background(), "a1", "u1", &invalidRole, nil)
	if err != domain.ErrInvalidRole {
		t.Fatal("expected invalid role err")
	}

	// Self demotion
	demotedRole := "learner"
	_, err = uc.Execute(context.Background(), "a1", "a1", &demotedRole, nil)
	if err != domain.ErrSelfDemotion {
		t.Fatal("expected self demotion err")
	}
}

func TestListAuditLog(t *testing.T) {
	auditRepo := &mockAuditRepo{
		records: []domain.AuditRecord{
			{ID: "audit-1"},
		},
	}
	uc := application.NewGetAuditLogUseCase(auditRepo)
	res, _ := uc.Execute(context.Background())
	if len(res) != 1 {
		t.Fatal("expected 1 record")
	}
}
