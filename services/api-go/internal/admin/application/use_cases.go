package application

import (
	"context"
	"encoding/json"

	"github.com/hcl-backend/services/api-go/internal/admin/domain"
)

type ListUsersUseCase struct {
	identitySvc IdentityService
}

func NewListUsersUseCase(identitySvc IdentityService) *ListUsersUseCase {
	return &ListUsersUseCase{identitySvc: identitySvc}
}

func (uc *ListUsersUseCase) Execute(ctx context.Context) ([]domain.User, error) {
	return uc.identitySvc.ListUsers(ctx)
}

type UpdateUserUseCase struct {
	identitySvc IdentityService
	auditRepo   AuditRepository
}

func NewUpdateUserUseCase(identitySvc IdentityService, auditRepo AuditRepository) *UpdateUserUseCase {
	return &UpdateUserUseCase{identitySvc: identitySvc, auditRepo: auditRepo}
}

func (uc *UpdateUserUseCase) Execute(ctx context.Context, adminID, targetID string, role, status *string) (*domain.User, error) {
	if role != nil && *role != "learner" && *role != "curator" && *role != "admin" {
		return nil, domain.ErrInvalidRole
	}
	if status != nil && *status != "active" && *status != "suspended" {
		return nil, domain.ErrInvalidStatus
	}

	if adminID == targetID {
		if role != nil && *role != "admin" {
			return nil, domain.ErrSelfDemotion
		}
		if status != nil && *status == "suspended" {
			return nil, domain.ErrSelfDemotion
		}
	}

	beforeUser, err := uc.identitySvc.GetUserByID(ctx, targetID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	afterUser, err := uc.identitySvc.UpdateUser(ctx, targetID, role, status)
	if err != nil {
		return nil, err
	}

	beforeState, _ := json.Marshal(beforeUser)
	afterState, _ := json.Marshal(afterUser)

	auditRecord := &domain.AuditRecord{
		ActorID:          adminID,
		Action:           "UPDATE_USER",
		TargetEntityType: "users",
		TargetEntityID:   targetID,
		BeforeState:      json.RawMessage(beforeState),
		AfterState:       json.RawMessage(afterState),
	}

	// We create the audit record. If it fails, we log it, but in a real system
	// this should ideally be in a transaction across Identity and Admin.
	// Since Identity owns users, we do our best here. If the architecture provided
	// a unified TxManager, we would use it.
	if err := uc.auditRepo.Create(ctx, auditRecord); err != nil {
		return nil, err
	}

	return afterUser, nil
}

type GetAuditLogUseCase struct {
	auditRepo AuditRepository
}

func NewGetAuditLogUseCase(auditRepo AuditRepository) *GetAuditLogUseCase {
	return &GetAuditLogUseCase{auditRepo: auditRepo}
}

func (uc *GetAuditLogUseCase) Execute(ctx context.Context) ([]domain.AuditRecord, error) {
	return uc.auditRepo.List(ctx)
}
