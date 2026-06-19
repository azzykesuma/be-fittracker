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

func (svc *Service) StartSession(ctx context.Context, userID string, req startSessionRequest) (workoutSessionResponse, error) {
	req.WorkoutPlanID = strings.TrimSpace(req.WorkoutPlanID)
	if req.WorkoutPlanID == "" {
		return workoutSessionResponse{}, errors.New("workout_plan_id is required")
	}
	exists, err := svc.repo.PlanBelongsToUser(ctx, req.WorkoutPlanID, userID)
	if err != nil {
		return workoutSessionResponse{}, err
	}
	if !exists {
		return workoutSessionResponse{}, errors.New("workout plan not found")
	}
	record, err := svc.repo.CreateSession(ctx, uuid.NewString(), userID, req.WorkoutPlanID, req.Notes)
	if err != nil {
		return workoutSessionResponse{}, err
	}
	return toSessionResponse(record), nil
}

func (svc *Service) ListSessions(ctx context.Context, userID string, filter listSessionsFilter) ([]workoutSessionResponse, error) {
	filter.Status = strings.TrimSpace(strings.ToLower(filter.Status))
	if filter.Status != "" && filter.Status != "in_progress" && filter.Status != "finished" && filter.Status != "cancelled" {
		return nil, errors.New("invalid status")
	}
	records, err := svc.repo.ListSessions(ctx, userID, filter.From, filter.To, filter.Status)
	if err != nil {
		return nil, err
	}
	responses := make([]workoutSessionResponse, 0, len(records))
	for _, r := range records {
		responses = append(responses, toSessionResponse(r))
	}
	return responses, nil
}

func (svc *Service) FindSession(ctx context.Context, id string, userID string) (workoutSessionDetailResponse, error) {
	record, err := svc.repo.FindSession(ctx, id, userID)
	if err != nil {
		return workoutSessionDetailResponse{}, err
	}
	sets, err := svc.repo.ListSetLogs(ctx, record.ID)
	if err != nil {
		return workoutSessionDetailResponse{}, err
	}
	return toSessionDetailResponse(record, toSetResponses(sets)), nil
}

func (svc *Service) LogSet(ctx context.Context, sessionID string, userID string, req logSetRequest) (workoutSetResponse, error) {
	_, err := svc.repo.FindSession(ctx, sessionID, userID)
	if err != nil {
		return workoutSetResponse{}, err
	}

	req.ExerciseName = strings.TrimSpace(req.ExerciseName)
	req.ExerciseID = strings.TrimSpace(req.ExerciseID)

	var exerciseName string
	if req.ExerciseName != "" {
		exerciseName = req.ExerciseName
	} else if req.ExerciseID != "" {
		err := svc.repo.db.QueryRow(ctx, `
			SELECT e.name 
			FROM exercises e
			JOIN workout_plans wp ON e.workout_plan_id = wp.id
			WHERE e.id = $1 AND wp.user_id = $2
		`, req.ExerciseID, userID).Scan(&exerciseName)
		if err != nil {
			exerciseName = "Unknown exercise"
		}
	} else {
		return workoutSetResponse{}, errors.New("exercise_name or exercise_id is required")
	}
	req.ExerciseName = exerciseName

	if req.SetNumber <= 0 {
		return workoutSetResponse{}, errors.New("set_number must be greater than 0")
	}
	if req.Reps < 0 {
		return workoutSetResponse{}, errors.New("reps must be non-negative")
	}
	if req.WeightKG != nil && *req.WeightKG < 0 {
		return workoutSetResponse{}, errors.New("weight_kg must be non-negative")
	}

	var exerciseIDPtr *string
	if req.ExerciseID != "" {
		exerciseIDPtr = &req.ExerciseID
	}

	record, err := svc.repo.CreateSetLog(ctx, uuid.NewString(), sessionID, exerciseIDPtr, req)
	if err != nil {
		return workoutSetResponse{}, err
	}
	return toSetResponse(record), nil
}

func (svc *Service) FinishSession(ctx context.Context, id string, userID string, req finishSessionRequest) (workoutSessionResponse, error) {
	_, err := svc.repo.FindSession(ctx, id, userID)
	if err != nil {
		return workoutSessionResponse{}, err
	}
	record, err := svc.repo.FinishSession(ctx, id, userID, req.Notes)
	if err != nil {
		return workoutSessionResponse{}, err
	}
	return toSessionResponse(record), nil
}

func (svc *Service) DeleteSession(ctx context.Context, id string, userID string) error {
	_, err := svc.repo.FindSession(ctx, id, userID)
	if err != nil {
		return err
	}
	return svc.repo.DeleteSession(ctx, id, userID)
}

func toSessionResponse(record workoutSessionRecord) workoutSessionResponse {
	return workoutSessionResponse{
		ID:              record.ID,
		WorkoutPlanID:   record.WorkoutPlanID,
		WorkoutPlanName: record.WorkoutPlanName,
		StartedAt:       record.StartedAt,
		FinishedAt:      record.FinishedAt,
		Status:          record.Status,
		Notes:           record.Notes,
	}
}

func toSessionDetailResponse(record workoutSessionRecord, sets []workoutSetResponse) workoutSessionDetailResponse {
	return workoutSessionDetailResponse{
		ID:              record.ID,
		WorkoutPlanID:   record.WorkoutPlanID,
		WorkoutPlanName: record.WorkoutPlanName,
		StartedAt:       record.StartedAt,
		FinishedAt:      record.FinishedAt,
		Status:          record.Status,
		Notes:           record.Notes,
		Sets:            sets,
	}
}

func toSetResponses(records []workoutSetRecord) []workoutSetResponse {
	responses := make([]workoutSetResponse, 0, len(records))
	for _, r := range records {
		responses = append(responses, toSetResponse(r))
	}
	return responses
}

func toSetResponse(record workoutSetRecord) workoutSetResponse {
	return workoutSetResponse{
		ID:           record.ID,
		ExerciseID:   record.ExerciseID,
		ExerciseName: record.ExerciseName,
		SetNumber:    record.SetNumber,
		Reps:         record.Reps,
		WeightKG:     record.WeightKG,
		Completed:    record.Completed,
		CreatedAt:    record.CreatedAt,
	}
}

