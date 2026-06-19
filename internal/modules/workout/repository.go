package workout

import (
	"context"
	"fmt"
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

type workoutSessionRecord struct {
	ID              string
	UserID          string
	WorkoutPlanID   *string
	WorkoutPlanName string
	StartedAt       time.Time
	FinishedAt      *time.Time
	Status          string
	Notes           string
	UpdatedAt       time.Time
}

type workoutSetRecord struct {
	ID               string
	WorkoutSessionID string
	ExerciseID       *string
	ExerciseName     string
	SetNumber        int
	Reps             int
	WeightKG         *float64
	Completed        bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (repo *Repository) CreateSession(ctx context.Context, id string, userID string, planID string, notes string) (workoutSessionRecord, error) {
	var record workoutSessionRecord
	err := repo.db.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO workout_sessions (id, user_id, workout_plan_id, notes)
			VALUES ($1, $2, $3, NULLIF($4, ''))
			RETURNING id, user_id, workout_plan_id, started_at, finished_at, status, notes, updated_at
		)
		SELECT i.id, i.user_id, i.workout_plan_id, COALESCE(wp.name, ''), i.started_at, i.finished_at, i.status, COALESCE(i.notes, ''), i.updated_at
		FROM inserted i
		LEFT JOIN workout_plans wp ON i.workout_plan_id = wp.id
	`, id, userID, planID, notes).Scan(
		&record.ID, &record.UserID, &record.WorkoutPlanID, &record.WorkoutPlanName,
		&record.StartedAt, &record.FinishedAt, &record.Status, &record.Notes, &record.UpdatedAt,
	)
	return record, err
}

func (repo *Repository) ListSessions(ctx context.Context, userID string, fromDate string, toDate string, status string) ([]workoutSessionRecord, error) {
	query := `
		SELECT ws.id, ws.user_id, ws.workout_plan_id, COALESCE(wp.name, 'Deleted workout plan'), ws.started_at, ws.finished_at, ws.status, COALESCE(ws.notes, ''), ws.updated_at
		FROM workout_sessions ws
		LEFT JOIN workout_plans wp ON ws.workout_plan_id = wp.id
		WHERE ws.user_id = $1
	`
	args := []any{userID}
	argIndex := 2

	if fromDate != "" {
		query += fmt.Sprintf(" AND ws.started_at >= $%d::timestamptz", argIndex)
		args = append(args, fromDate)
		argIndex++
	}
	if toDate != "" {
		query += fmt.Sprintf(" AND ws.started_at < ($%d::date + 1)::timestamptz", argIndex)
		args = append(args, toDate)
		argIndex++
	}
	if status != "" {
		query += fmt.Sprintf(" AND ws.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += " ORDER BY ws.started_at DESC"

	rows, err := repo.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []workoutSessionRecord{}
	for rows.Next() {
		var rec workoutSessionRecord
		err := rows.Scan(
			&rec.ID, &rec.UserID, &rec.WorkoutPlanID, &rec.WorkoutPlanName,
			&rec.StartedAt, &rec.FinishedAt, &rec.Status, &rec.Notes, &rec.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (repo *Repository) FindSession(ctx context.Context, id string, userID string) (workoutSessionRecord, error) {
	var record workoutSessionRecord
	err := repo.db.QueryRow(ctx, `
		SELECT ws.id, ws.user_id, ws.workout_plan_id, COALESCE(wp.name, 'Deleted workout plan'), ws.started_at, ws.finished_at, ws.status, COALESCE(ws.notes, ''), ws.updated_at
		FROM workout_sessions ws
		LEFT JOIN workout_plans wp ON ws.workout_plan_id = wp.id
		WHERE ws.id = $1 AND ws.user_id = $2
	`, id, userID).Scan(
		&record.ID, &record.UserID, &record.WorkoutPlanID, &record.WorkoutPlanName,
		&record.StartedAt, &record.FinishedAt, &record.Status, &record.Notes, &record.UpdatedAt,
	)
	return record, err
}

func (repo *Repository) CreateSetLog(ctx context.Context, id string, sessionID string, exerciseID *string, req logSetRequest) (workoutSetRecord, error) {
	var record workoutSetRecord
	completed := true
	if req.Completed != nil {
		completed = *req.Completed
	}
	err := repo.db.QueryRow(ctx, `
		INSERT INTO workout_set_logs (id, workout_session_id, exercise_id, exercise_name, set_number, reps, weight_kg, completed)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, workout_session_id, exercise_id, exercise_name, set_number, reps, weight_kg::float8, completed, created_at, updated_at
	`, id, sessionID, exerciseID, req.ExerciseName, req.SetNumber, req.Reps, req.WeightKG, completed).Scan(
		&record.ID, &record.WorkoutSessionID, &record.ExerciseID, &record.ExerciseName,
		&record.SetNumber, &record.Reps, &record.WeightKG, &record.Completed, &record.CreatedAt, &record.UpdatedAt,
	)
	return record, err
}

func (repo *Repository) ListSetLogs(ctx context.Context, sessionID string) ([]workoutSetRecord, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT id, workout_session_id, exercise_id, exercise_name, set_number, reps, weight_kg::float8, completed, created_at, updated_at
		FROM workout_set_logs
		WHERE workout_session_id = $1
		ORDER BY created_at ASC, set_number ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []workoutSetRecord{}
	for rows.Next() {
		var record workoutSetRecord
		err := rows.Scan(
			&record.ID, &record.WorkoutSessionID, &record.ExerciseID, &record.ExerciseName,
			&record.SetNumber, &record.Reps, &record.WeightKG, &record.Completed, &record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (repo *Repository) FinishSession(ctx context.Context, id string, userID string, notes string) (workoutSessionRecord, error) {
	var record workoutSessionRecord
	err := repo.db.QueryRow(ctx, `
		WITH updated AS (
			UPDATE workout_sessions
			SET status = 'finished', finished_at = NOW(), notes = COALESCE(NULLIF($3, ''), notes), updated_at = NOW()
			WHERE id = $1 AND user_id = $2
			RETURNING id, user_id, workout_plan_id, started_at, finished_at, status, notes, updated_at
		)
		SELECT u.id, u.user_id, u.workout_plan_id, COALESCE(wp.name, 'Deleted workout plan'), u.started_at, u.finished_at, u.status, COALESCE(u.notes, ''), u.updated_at
		FROM updated u
		LEFT JOIN workout_plans wp ON u.workout_plan_id = wp.id
	`, id, userID, notes).Scan(
		&record.ID, &record.UserID, &record.WorkoutPlanID, &record.WorkoutPlanName,
		&record.StartedAt, &record.FinishedAt, &record.Status, &record.Notes, &record.UpdatedAt,
	)
	return record, err
}

func (repo *Repository) DeleteSession(ctx context.Context, id string, userID string) error {
	_, err := repo.db.Exec(ctx, `DELETE FROM workout_sessions WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

