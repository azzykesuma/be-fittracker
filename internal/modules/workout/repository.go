package workout

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"be-fittracker/internal/database"
)

type Repository struct {
	db database.Querier
}

func NewRepository(db database.Querier) *Repository {
	return &Repository{db: db}
}

type workoutPlanRecord struct {
	ID           string
	Name         string
	Description  string
	ScheduledDay string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type exerciseRecord struct {
	ID             string
	WorkoutPlanID  string
	Name           string
	MuscleGroup    string
	TargetSets     *int
	TargetReps     *int
	TargetWeightKG *float64
	RestSeconds    int
	OrderIndex     int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (repo *Repository) CreatePlan(ctx context.Context, id string, userID string, req workoutPlanRequest) (workoutPlanRecord, error) {
	var record workoutPlanRecord
	err := repo.db.QueryRow(ctx, `
		INSERT INTO workout_plans (id, user_id, name, description, scheduled_day)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''))
		RETURNING id, name, COALESCE(description, ''), COALESCE(scheduled_day, ''), created_at, updated_at
	`, id, userID, req.Name, req.Description, req.ScheduledDay).Scan(&record.ID, &record.Name, &record.Description, &record.ScheduledDay, &record.CreatedAt, &record.UpdatedAt)
	return record, err
}

func (repo *Repository) ListPlans(ctx context.Context, userID string) ([]workoutPlanRecord, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT id, name, COALESCE(description, ''), COALESCE(scheduled_day, ''), created_at, updated_at
		FROM workout_plans
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []workoutPlanRecord{}
	for rows.Next() {
		var record workoutPlanRecord
		if err := rows.Scan(&record.ID, &record.Name, &record.Description, &record.ScheduledDay, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, record)
	}
	return items, rows.Err()
}

func (repo *Repository) ListPlansByDay(ctx context.Context, userID string, scheduledDay string) ([]workoutPlanRecord, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT id, name, COALESCE(description, ''), COALESCE(scheduled_day, ''), created_at, updated_at
		FROM workout_plans
		WHERE user_id = $1 AND scheduled_day = $2
		ORDER BY created_at DESC
	`, userID, scheduledDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []workoutPlanRecord{}
	for rows.Next() {
		var record workoutPlanRecord
		if err := rows.Scan(&record.ID, &record.Name, &record.Description, &record.ScheduledDay, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, record)
	}
	return items, rows.Err()
}

func (repo *Repository) FindPlan(ctx context.Context, id string, userID string) (workoutPlanRecord, error) {
	var record workoutPlanRecord
	err := repo.db.QueryRow(ctx, `
		SELECT id, name, COALESCE(description, ''), COALESCE(scheduled_day, ''), created_at, updated_at
		FROM workout_plans
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&record.ID, &record.Name, &record.Description, &record.ScheduledDay, &record.CreatedAt, &record.UpdatedAt)
	return record, err
}

func (repo *Repository) UpdatePlan(ctx context.Context, id string, userID string, req workoutPlanRequest) (workoutPlanRecord, error) {
	var record workoutPlanRecord
	err := repo.db.QueryRow(ctx, `
		UPDATE workout_plans
		SET name = $3, description = NULLIF($4, ''), scheduled_day = NULLIF($5, ''), updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, name, COALESCE(description, ''), COALESCE(scheduled_day, ''), created_at, updated_at
	`, id, userID, req.Name, req.Description, req.ScheduledDay).Scan(&record.ID, &record.Name, &record.Description, &record.ScheduledDay, &record.CreatedAt, &record.UpdatedAt)
	return record, err
}

func (repo *Repository) DeletePlan(ctx context.Context, id string, userID string) error {
	_, err := repo.db.Exec(ctx, `DELETE FROM workout_plans WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (repo *Repository) PlanBelongsToUser(ctx context.Context, id string, userID string) (bool, error) {
	var exists bool
	err := repo.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM workout_plans WHERE id = $1 AND user_id = $2)`, id, userID).Scan(&exists)
	return exists, err
}

func (repo *Repository) CreateExercise(ctx context.Context, id string, planID string, req exerciseRequest) (exerciseRecord, error) {
	return scanExercise(repo.db.QueryRow(ctx, `
		INSERT INTO exercises (id, workout_plan_id, name, muscle_group, target_sets, target_reps, target_weight_kg, rest_seconds, order_index)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7, $8, $9)
		RETURNING id, workout_plan_id, name, COALESCE(muscle_group, ''), target_sets, target_reps, target_weight_kg::float8, rest_seconds, order_index, created_at, updated_at
	`, id, planID, req.Name, req.MuscleGroup, req.TargetSets, req.TargetReps, req.TargetWeightKG, valueOrDefault(req.RestSeconds, 60), valueOrDefault(req.OrderIndex, 0)))
}

func (repo *Repository) ListExercises(ctx context.Context, planID string) ([]exerciseRecord, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT id, workout_plan_id, name, COALESCE(muscle_group, ''), target_sets, target_reps, target_weight_kg::float8, rest_seconds, order_index, created_at, updated_at
		FROM exercises
		WHERE workout_plan_id = $1
		ORDER BY order_index ASC, created_at ASC
	`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []exerciseRecord{}
	for rows.Next() {
		record, err := scanExercise(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, record)
	}
	return items, rows.Err()
}

func (repo *Repository) UpdateExercise(ctx context.Context, id string, userID string, req exerciseRequest) (exerciseRecord, error) {
	return scanExercise(repo.db.QueryRow(ctx, `
		UPDATE exercises e
		SET name = $3,
		    muscle_group = NULLIF($4, ''),
		    target_sets = $5,
		    target_reps = $6,
		    target_weight_kg = $7,
		    rest_seconds = $8,
		    order_index = $9,
		    updated_at = NOW()
		FROM workout_plans wp
		WHERE e.id = $1 AND e.workout_plan_id = wp.id AND wp.user_id = $2
		RETURNING e.id, e.workout_plan_id, e.name, COALESCE(e.muscle_group, ''), e.target_sets, e.target_reps, e.target_weight_kg::float8, e.rest_seconds, e.order_index, e.created_at, e.updated_at
	`, id, userID, req.Name, req.MuscleGroup, req.TargetSets, req.TargetReps, req.TargetWeightKG, valueOrDefault(req.RestSeconds, 60), valueOrDefault(req.OrderIndex, 0)))
}

func (repo *Repository) DeleteExercise(ctx context.Context, id string, userID string) error {
	_, err := repo.db.Exec(ctx, `
		DELETE FROM exercises e
		USING workout_plans wp
		WHERE e.id = $1 AND e.workout_plan_id = wp.id AND wp.user_id = $2
	`, id, userID)
	return err
}

func valueOrDefault(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

type exerciseScanner interface {
	Scan(dest ...any) error
}

func scanExercise(scanner exerciseScanner) (exerciseRecord, error) {
	var record exerciseRecord
	var targetSets pgtype.Int4
	var targetReps pgtype.Int4
	var targetWeight pgtype.Float8
	err := scanner.Scan(&record.ID, &record.WorkoutPlanID, &record.Name, &record.MuscleGroup, &targetSets, &targetReps, &targetWeight, &record.RestSeconds, &record.OrderIndex, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return exerciseRecord{}, err
	}
	if targetSets.Valid {
		value := int(targetSets.Int32)
		record.TargetSets = &value
	}
	if targetReps.Valid {
		value := int(targetReps.Int32)
		record.TargetReps = &value
	}
	if targetWeight.Valid {
		value := targetWeight.Float64
		record.TargetWeightKG = &value
	}
	return record, nil
}

var _ exerciseScanner = pgx.Row(nil)
