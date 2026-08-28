package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/hcl-backend/services/api-go/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

// Mock TxManager to test the interface without a real database
type mockTxManager struct {
	called bool
}

func (m *mockTxManager) Do(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	m.called = true
	return fn(ctx, nil) // passing nil tx for test
}

func TestTxManagerInterface(t *testing.T) {
	tm := &mockTxManager{}

	err := tm.Do(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !tm.called {
		t.Error("Expected TxManager.Do to be called")
	}
}

func TestNewPool_InvalidDSN(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use an invalid DSN to ensure ParseConfig fails or Ping fails
	_, err := database.NewPool(ctx, "postgres://invalid:invalid@localhost:1/invalid?sslmode=disable")
	if err == nil {
		t.Error("Expected error for invalid DSN, got nil")
	}
}
