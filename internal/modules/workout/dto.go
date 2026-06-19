package workout

import "time"

type workoutPlanRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ScheduledDay string `json:"scheduled_day"`
}

type exerciseRequest struct {
	Name           string   `json:"name"`
	MuscleGroup    string   `json:"muscle_group"`
	TargetSets     *int     `json:"target_sets"`
	TargetReps     *int     `json:"target_reps"`
	TargetWeightKG *float64 `json:"target_weight_kg"`
	RestSeconds    *int     `json:"rest_seconds"`
	OrderIndex     *int     `json:"order_index"`
}

type workoutPlanResponse struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	ScheduledDay string             `json:"scheduled_day,omitempty"`
	Exercises    []exerciseResponse `json:"exercises,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

type exerciseResponse struct {
	ID             string    `json:"id"`
	WorkoutPlanID  string    `json:"workout_plan_id"`
	Name           string    `json:"name"`
	MuscleGroup    string    `json:"muscle_group,omitempty"`
	TargetSets     *int      `json:"target_sets,omitempty"`
	TargetReps     *int      `json:"target_reps,omitempty"`
	TargetWeightKG *float64  `json:"target_weight_kg,omitempty"`
	RestSeconds    int       `json:"rest_seconds"`
	OrderIndex     int       `json:"order_index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type startSessionRequest struct {
	WorkoutPlanID string `json:"workout_plan_id"`
	Notes         string `json:"notes"`
}

type listSessionsFilter struct {
	From   string
	To     string
	Status string
}

type workoutSessionResponse struct {
	ID              string     `json:"id"`
	WorkoutPlanID   *string    `json:"workout_plan_id,omitempty"`
	WorkoutPlanName string     `json:"workout_plan_name,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	Status          string     `json:"status"`
	Notes           string     `json:"notes,omitempty"`
}

type workoutSessionDetailResponse struct {
	ID              string               `json:"id"`
	WorkoutPlanID   *string              `json:"workout_plan_id,omitempty"`
	WorkoutPlanName string               `json:"workout_plan_name,omitempty"`
	StartedAt       time.Time            `json:"started_at"`
	FinishedAt      *time.Time           `json:"finished_at,omitempty"`
	Status          string               `json:"status"`
	Notes           string               `json:"notes,omitempty"`
	Sets            []workoutSetResponse `json:"sets"`
}

type logSetRequest struct {
	ExerciseID   string   `json:"exercise_id"`
	ExerciseName string   `json:"exercise_name"`
	SetNumber    int      `json:"set_number"`
	Reps         int      `json:"reps"`
	WeightKG     *float64 `json:"weight_kg"`
	Completed    *bool    `json:"completed"`
}

type workoutSetResponse struct {
	ID           string    `json:"id"`
	ExerciseID   *string   `json:"exercise_id,omitempty"`
	ExerciseName string    `json:"exercise_name"`
	SetNumber    int       `json:"set_number"`
	Reps         int       `json:"reps"`
	WeightKG     *float64  `json:"weight_kg,omitempty"`
	Completed    bool      `json:"completed"`
	CreatedAt    time.Time `json:"created_at"`
}

type finishSessionRequest struct {
	Notes string `json:"notes"`
}

