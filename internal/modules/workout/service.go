package workout

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (svc *Service) CreatePlan(ctx context.Context, userID string, req workoutPlanRequest) (workoutPlanResponse, error) {
	req, err := validatePlanRequest(req)
	if err != nil {
		return workoutPlanResponse{}, err
	}
	record, err := svc.repo.CreatePlan(ctx, uuid.NewString(), userID, req)
	if err != nil {
		return workoutPlanResponse{}, err
	}
	return toPlanResponse(record, nil), nil
}

func (svc *Service) ListPlans(ctx context.Context, userID string) ([]workoutPlanResponse, error) {
	records, err := svc.repo.ListPlans(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]workoutPlanResponse, 0, len(records))
	for _, record := range records {
		items = append(items, toPlanResponse(record, nil))
	}
	return items, nil
}

func (svc *Service) TodayPlans(ctx context.Context, userID string) ([]workoutPlanResponse, error) {
	day := strings.ToLower(time.Now().Weekday().String())
	records, err := svc.repo.ListPlansByDay(ctx, userID, day)
	if err != nil {
		return nil, err
	}
	items := make([]workoutPlanResponse, 0, len(records))
	for _, record := range records {
		exercises, err := svc.repo.ListExercises(ctx, record.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, toPlanResponse(record, toExerciseResponses(exercises)))
	}
	return items, nil
}

func (svc *Service) FindPlan(ctx context.Context, id string, userID string) (workoutPlanResponse, error) {
	record, err := svc.repo.FindPlan(ctx, id, userID)
	if err != nil {
		return workoutPlanResponse{}, err
	}
	exercises, err := svc.repo.ListExercises(ctx, record.ID)
	if err != nil {
		return workoutPlanResponse{}, err
	}
	return toPlanResponse(record, toExerciseResponses(exercises)), nil
}

func (svc *Service) UpdatePlan(ctx context.Context, id string, userID string, req workoutPlanRequest) (workoutPlanResponse, error) {
	req, err := validatePlanRequest(req)
	if err != nil {
		return workoutPlanResponse{}, err
	}
	record, err := svc.repo.UpdatePlan(ctx, id, userID, req)
	if err != nil {
		return workoutPlanResponse{}, err
	}
	return toPlanResponse(record, nil), nil
}

func (svc *Service) DeletePlan(ctx context.Context, id string, userID string) error {
	return svc.repo.DeletePlan(ctx, id, userID)
}

func (svc *Service) CreateExercise(ctx context.Context, planID string, userID string, req exerciseRequest) (exerciseResponse, error) {
	exists, err := svc.repo.PlanBelongsToUser(ctx, planID, userID)
	if err != nil {
		return exerciseResponse{}, err
	}
	if !exists {
		return exerciseResponse{}, errors.New("workout plan not found")
	}
	req, err = validateExerciseRequest(req)
	if err != nil {
		return exerciseResponse{}, err
	}
	record, err := svc.repo.CreateExercise(ctx, uuid.NewString(), planID, req)
	if err != nil {
		return exerciseResponse{}, err
	}
	return toExerciseResponse(record), nil
}

func (svc *Service) UpdateExercise(ctx context.Context, id string, userID string, req exerciseRequest) (exerciseResponse, error) {
	req, err := validateExerciseRequest(req)
	if err != nil {
		return exerciseResponse{}, err
	}
	record, err := svc.repo.UpdateExercise(ctx, id, userID, req)
	if err != nil {
		return exerciseResponse{}, err
	}
	return toExerciseResponse(record), nil
}

func (svc *Service) DeleteExercise(ctx context.Context, id string, userID string) error {
	return svc.repo.DeleteExercise(ctx, id, userID)
}

func validatePlanRequest(req workoutPlanRequest) (workoutPlanRequest, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.ScheduledDay = strings.ToLower(strings.TrimSpace(req.ScheduledDay))
	if req.Name == "" {
		return req, errors.New("name is required")
	}
	if req.ScheduledDay != "" && !validScheduledDay(req.ScheduledDay) {
		return req, errors.New("scheduled_day must be a weekday")
	}
	return req, nil
}

func validateExerciseRequest(req exerciseRequest) (exerciseRequest, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.MuscleGroup = strings.TrimSpace(req.MuscleGroup)
	if req.Name == "" {
		return req, errors.New("name is required")
	}
	if req.TargetSets != nil && *req.TargetSets <= 0 {
		return req, errors.New("target_sets must be greater than 0")
	}
	if req.TargetReps != nil && *req.TargetReps <= 0 {
		return req, errors.New("target_reps must be greater than 0")
	}
	if req.TargetWeightKG != nil && *req.TargetWeightKG < 0 {
		return req, errors.New("target_weight_kg must be non-negative")
	}
	if req.RestSeconds != nil && *req.RestSeconds < 0 {
		return req, errors.New("rest_seconds must be non-negative")
	}
	if req.OrderIndex != nil && *req.OrderIndex < 0 {
		return req, errors.New("order_index must be non-negative")
	}
	return req, nil
}

func validScheduledDay(day string) bool {
	switch day {
	case "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday":
		return true
	default:
		return false
	}
}

func toPlanResponse(record workoutPlanRecord, exercises []exerciseResponse) workoutPlanResponse {
	return workoutPlanResponse{ID: record.ID, Name: record.Name, Description: record.Description, ScheduledDay: record.ScheduledDay, Exercises: exercises, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}
}

func toExerciseResponses(records []exerciseRecord) []exerciseResponse {
	items := make([]exerciseResponse, 0, len(records))
	for _, record := range records {
		items = append(items, toExerciseResponse(record))
	}
	return items
}

func toExerciseResponse(record exerciseRecord) exerciseResponse {
	return exerciseResponse{ID: record.ID, WorkoutPlanID: record.WorkoutPlanID, Name: record.Name, MuscleGroup: record.MuscleGroup, TargetSets: record.TargetSets, TargetReps: record.TargetReps, TargetWeightKG: record.TargetWeightKG, RestSeconds: record.RestSeconds, OrderIndex: record.OrderIndex, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}
}
