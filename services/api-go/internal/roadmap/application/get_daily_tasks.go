package application

import (
	"context"
	"time"

	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
)

type GetDailyTasksUseCase struct {
	repo RoadmapRepository
}

func NewGetDailyTasksUseCase(repo RoadmapRepository) *GetDailyTasksUseCase {
	return &GetDailyTasksUseCase{repo: repo}
}

func (uc *GetDailyTasksUseCase) Execute(ctx context.Context, learnerID string) ([]domain.DailyTaskDay, error) {
	now := time.Now()
	// ISO week start: Monday
	offset := int(now.Weekday()) - 1
	if offset < 0 {
		offset = 6 // Sunday
	}
	monday := now.AddDate(0, 0, -offset)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, now.Location())
	end := monday.AddDate(0, 0, 6)

	return uc.repo.GetDailyTasks(ctx, learnerID, monday, end)
}
