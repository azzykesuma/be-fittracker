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
