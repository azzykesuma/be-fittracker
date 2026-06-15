package habit

import "time"

type createHabitRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Frequency   string `json:"frequency"`
	TargetCount int    `json:"target_count"`
}

type habitResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Frequency      string    `json:"frequency"`
	TargetCount    int       `json:"target_count"`
	IsActive       bool      `json:"is_active"`
	CompletedToday bool      `json:"completed_today"`
	CurrentStreak  int       `json:"current_streak"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
