package report

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"be-fittracker/internal/database"
)

type Repository struct {
	db database.Querier
}

func NewRepository(db database.Querier) *Repository {
	return &Repository{db: db}
}

type UserSummaryRecord struct {
	Name             string
	Email            string
	HeightCM         *int
	WeightKG         *float64
	BMI              *float64
	BodyFat          *float64
	TotalWorkouts    int
	AvgDailyCalories float64
	CreatedAt        time.Time
}

type ProgressRecord struct {
	LogDate           time.Time
	WeightKG          float64
	BMI               float64
	BodyFatPercentage float64
	NeckCM            float64
	ShoulderCM        float64
	ChestCM           float64
	WaistCM           float64
	BellyCM           float64
	HipsCM            float64
	LeftBicepCM       float64
	RightBicepCM      float64
	LeftForearmCM     float64
	RightForearmCM    float64
	LeftThighCM       float64
	RightThighCM      float64
	LeftCalfCM        float64
	RightCalfCM       float64
	Notes             string
}

type MealRecord struct {
	MealDate time.Time
	MealType string
	FoodName string
	Calories int
	ProteinG float64
	CarbsG   float64
	FatG     float64
	Notes    string
}

type WorkoutRecord struct {
	StartedAt    time.Time
	PlanName     string
	Status       string
	SessionNotes string
	ExerciseName string
	SetNumber    int
	Reps         int
	WeightKG     float64
	Completed    bool
}

func (repo *Repository) GetUserSummary(ctx context.Context, userID string) (UserSummaryRecord, error) {
	var summary UserSummaryRecord

	var height sql.NullInt32
	var weight sql.NullFloat64
	err := repo.db.QueryRow(ctx, `
		SELECT name, email, height_cm, weight_kg::float8, created_at
		FROM users
		WHERE id = $1
	`, userID).Scan(&summary.Name, &summary.Email, &height, &weight, &summary.CreatedAt)
	if err != nil {
		return summary, err
	}

	if height.Valid {
		val := int(height.Int32)
		summary.HeightCM = &val
	}
	if weight.Valid {
		summary.WeightKG = &weight.Float64
	}

	// Get latest measurements from log (weight, BMI, body fat)
	var latestWeight sql.NullFloat64
	var latestBMI sql.NullFloat64
	var latestBodyFat sql.NullFloat64
	err = repo.db.QueryRow(ctx, `
		SELECT weight_kg::float8, bmi::float8, body_fat_percentage::float8
		FROM body_measurement_logs
		WHERE user_id = $1
		ORDER BY log_date DESC
		LIMIT 1
	`, userID).Scan(&latestWeight, &latestBMI, &latestBodyFat)
	if err == nil {
		if latestWeight.Valid {
			summary.WeightKG = &latestWeight.Float64
		}
		if latestBMI.Valid {
			summary.BMI = &latestBMI.Float64
		}
		if latestBodyFat.Valid {
			summary.BodyFat = &latestBodyFat.Float64
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return summary, err
	}

	// Get total workout sessions logged (finished)
	err = repo.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM workout_sessions
		WHERE user_id = $1 AND status = 'finished'
	`, userID).Scan(&summary.TotalWorkouts)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return summary, err
	}

	// Get average daily calories
	err = repo.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(daily_sum), 0.0)
		FROM (
			SELECT SUM(calories) as daily_sum
			FROM meal_logs
			WHERE user_id = $1
			GROUP BY meal_date
		) as sub
	`, userID).Scan(&summary.AvgDailyCalories)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return summary, err
	}

	return summary, nil
}

func (repo *Repository) GetProgressLogs(ctx context.Context, userID string) ([]ProgressRecord, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT
			log_date,
			weight_kg::float8,
			COALESCE(bmi::float8, 0.0),
			COALESCE(body_fat_percentage::float8, 0.0),
			COALESCE(neck_cm::float8, 0.0),
			COALESCE(shoulder_cm::float8, 0.0),
			COALESCE(chest_cm::float8, 0.0),
			COALESCE(waist_cm::float8, 0.0),
			COALESCE(belly_cm::float8, 0.0),
			COALESCE(hips_cm::float8, 0.0),
			COALESCE(left_bicep_cm::float8, 0.0),
			COALESCE(right_bicep_cm::float8, 0.0),
			COALESCE(left_forearm_cm::float8, 0.0),
			COALESCE(right_forearm_cm::float8, 0.0),
			COALESCE(left_thigh_cm::float8, 0.0),
			COALESCE(right_thigh_cm::float8, 0.0),
			COALESCE(left_calf_cm::float8, 0.0),
			COALESCE(right_calf_cm::float8, 0.0),
			COALESCE(notes, '')
		FROM body_measurement_logs
		WHERE user_id = $1
		ORDER BY log_date ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ProgressRecord
	for rows.Next() {
		var rec ProgressRecord
		err := rows.Scan(
			&rec.LogDate, &rec.WeightKG, &rec.BMI, &rec.BodyFatPercentage,
			&rec.NeckCM, &rec.ShoulderCM, &rec.ChestCM, &rec.WaistCM, &rec.BellyCM, &rec.HipsCM,
			&rec.LeftBicepCM, &rec.RightBicepCM, &rec.LeftForearmCM, &rec.RightForearmCM,
			&rec.LeftThighCM, &rec.RightThighCM, &rec.LeftCalfCM, &rec.RightCalfCM, &rec.Notes,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func (repo *Repository) GetMealLogs(ctx context.Context, userID string) ([]MealRecord, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT
			meal_date,
			meal_type,
			food_name,
			calories,
			COALESCE(protein_g::float8, 0.0),
			COALESCE(carbs_g::float8, 0.0),
			COALESCE(fat_g::float8, 0.0),
			COALESCE(notes, '')
		FROM meal_logs
		WHERE user_id = $1
		ORDER BY meal_date ASC, created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []MealRecord
	for rows.Next() {
		var rec MealRecord
		err := rows.Scan(&rec.MealDate, &rec.MealType, &rec.FoodName, &rec.Calories, &rec.ProteinG, &rec.CarbsG, &rec.FatG, &rec.Notes)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func (repo *Repository) GetWorkoutLogs(ctx context.Context, userID string) ([]WorkoutRecord, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT
			ws.started_at,
			COALESCE(wp.name, 'Custom / Deleted Plan'),
			ws.status,
			COALESCE(ws.notes, ''),
			wsl.exercise_name,
			wsl.set_number,
			wsl.reps,
			COALESCE(wsl.weight_kg::float8, 0.0),
			wsl.completed
		FROM workout_sessions ws
		LEFT JOIN workout_plans wp ON ws.workout_plan_id = wp.id
		JOIN workout_set_logs wsl ON wsl.workout_session_id = ws.id
		WHERE ws.user_id = $1
		ORDER BY ws.started_at ASC, wsl.created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []WorkoutRecord
	for rows.Next() {
		var rec WorkoutRecord
		err := rows.Scan(&rec.StartedAt, &rec.PlanName, &rec.Status, &rec.SessionNotes, &rec.ExerciseName, &rec.SetNumber, &rec.Reps, &rec.WeightKG, &rec.Completed)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}
