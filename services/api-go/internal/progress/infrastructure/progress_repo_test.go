package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hcl-backend/services/api-go/internal/progress/domain"
	"github.com/hcl-backend/services/api-go/internal/progress/infrastructure"
)

func TestRecordEngagement_SequentialPipeline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn := "postgres://postgres:postgres@localhost:5433/platform?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skip("Postgres unavailable for integration test")
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skip("Postgres ping failed")
		return
	}

	repo := infrastructure.NewPostgresProgressRepository(pool)
	learnerID := "de328d39-6901-4134-95e5-5adcf4801caa"

	// Reset target test path items to initial clean state
	_, _ = pool.Exec(ctx, `
		DELETE FROM platform.engagement_events
		WHERE learner_id = 'de328d39-6901-4134-95e5-5adcf4801caa';
		UPDATE platform.path_items
		SET state = CASE WHEN sequence_order = 1 THEN 'available' ELSE 'locked' END
		WHERE path_id = '9a223da3-b7da-446d-bc98-e45e2be86f54';
	`)

	// 1. Attempt out-of-order completion for Milestone 6 -> MUST BE REJECTED
	eventMod6 := &domain.EngagementEvent{
		ID:        uuid.New().String(),
		LearnerID: learnerID,
		ConceptID: "ai_eng_11_ai_system_design",
		EventType: "marked_reviewed",
		Timestamp: time.Now().UTC(),
	}
	err = repo.RecordEngagement(ctx, eventMod6)
	if err != domain.ErrPrerequisiteNotMet {
		t.Fatalf("Step 1 failed: Expected ErrPrerequisiteNotMet for Milestone 6 out-of-order completion, got: %v", err)
	}

	// 2. Complete Milestone 1 ("ai_eng_06_model_evaluation_and_experimentation") -> MUST SUCCEED
	eventMod1 := &domain.EngagementEvent{
		ID:        uuid.New().String(),
		LearnerID: learnerID,
		ConceptID: "ai_eng_06_model_evaluation_and_experimentation",
		EventType: "marked_reviewed",
		Timestamp: time.Now().UTC(),
	}
	err = repo.RecordEngagement(ctx, eventMod1)
	if err != nil {
		t.Fatalf("Step 2 failed: Expected clean completion for Milestone 1, got: %v", err)
	}

	// 3. Attempt out-of-order completion for Milestone 3 ("ai_eng_08_nlp_and_llms") -> MUST BE REJECTED
	eventMod3 := &domain.EngagementEvent{
		ID:        uuid.New().String(),
		LearnerID: learnerID,
		ConceptID: "ai_eng_08_nlp_and_llms",
		EventType: "marked_reviewed",
		Timestamp: time.Now().UTC(),
	}
	err = repo.RecordEngagement(ctx, eventMod3)
	if err != domain.ErrPrerequisiteNotMet {
		t.Fatalf("Step 3 failed: Expected ErrPrerequisiteNotMet for Milestone 3 out-of-order completion, got: %v", err)
	}

	// 4. Verify Summary reflects Milestone 1 as Competent (1 / 7 = 14%) and Milestone 2 as In Progress / Available
	summary, err := repo.Summary(ctx, learnerID, "0aabc4cc-0661-56ac-8b90-d9871a8d4df3")
	if err != nil {
		t.Fatalf("Step 4 failed: Summary error: %v", err)
	}

	if summary.TotalConcepts != 7 {
		t.Errorf("Expected 7 total concepts in summary, got: %d", summary.TotalConcepts)
	}
	if summary.Competent != 1 {
		t.Errorf("Expected 1 competent concept in summary, got: %d", summary.Competent)
	}
	if len(summary.ActivityData) == 0 {
		t.Errorf("Expected non-empty activityData array spanning 365 days")
	}

	t.Logf("Summary Trace: TotalConcepts=%d Competent=%d ActivityDays=%d BreakdownLen=%d",
		summary.TotalConcepts, summary.Competent, len(summary.ActivityData), len(summary.Breakdown))
}
