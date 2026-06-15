package habit

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (svc *Service) List(ctx context.Context, userID string) ([]habitResponse, error) {
	records, err := svc.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]habitResponse, 0, len(records))
	for _, record := range records {
		items = append(items, toHabitResponse(record))
	}
	return items, nil
}

func (svc *Service) Create(ctx context.Context, userID string, req createHabitRequest) (habitResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return habitResponse{}, errors.New("habit name is required")
	}

	frequency := strings.TrimSpace(req.Frequency)
	if frequency == "" {
		frequency = "daily"
	}

	targetCount := req.TargetCount
	if targetCount == 0 {
		targetCount = 1
	}
	if targetCount < 0 {
		return habitResponse{}, errors.New("target_count must be positive")
	}

	record, err := svc.repo.Create(ctx, uuid.NewString(), userID, name, strings.TrimSpace(req.Description), frequency, targetCount)
	if err != nil {
		return habitResponse{}, err
	}

	return toHabitResponse(record), nil
}

func (svc *Service) CompleteToday(ctx context.Context, habitID, userID string) error {
	exists, err := svc.repo.BelongsToUser(ctx, habitID, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("habit not found")
	}
	return svc.repo.CompleteToday(ctx, uuid.NewString(), habitID, userID)
}

func (svc *Service) UncompleteToday(ctx context.Context, habitID, userID string) error {
	exists, err := svc.repo.BelongsToUser(ctx, habitID, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("habit not found")
	}
	return svc.repo.UncompleteToday(ctx, habitID, userID)
}

func toHabitResponse(record habitRecord) habitResponse {
	return habitResponse{
		ID:             record.ID,
		Name:           record.Name,
		Description:    record.Description,
		Frequency:      record.Frequency,
		TargetCount:    record.TargetCount,
		IsActive:       record.IsActive,
		CompletedToday: record.CompletedToday,
		CurrentStreak:  record.CurrentStreak,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
	}
}
