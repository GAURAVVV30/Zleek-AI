package application

import (
	"context"
)

type ToggleDailyTaskUseCase struct {
	repo RoadmapRepository
}

func NewToggleDailyTaskUseCase(repo RoadmapRepository) *ToggleDailyTaskUseCase {
	return &ToggleDailyTaskUseCase{repo: repo}
}

func (uc *ToggleDailyTaskUseCase) Execute(ctx context.Context, learnerID, taskID string, completed bool) error {
	return uc.repo.ToggleDailyTask(ctx, learnerID, taskID, completed)
}
