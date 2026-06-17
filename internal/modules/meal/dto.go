package meal

import "time"

type mealLogRequest struct {
	MealDate string   `json:"meal_date"`
	MealType string   `json:"meal_type"`
	FoodName string   `json:"food_name"`
	Calories int      `json:"calories"`
	ProteinG *float64 `json:"protein_g"`
	CarbsG   *float64 `json:"carbs_g"`
	FatG     *float64 `json:"fat_g"`
	Notes    string   `json:"notes"`
}

type mealLogResponse struct {
	ID        string    `json:"id"`
	MealDate  string    `json:"meal_date"`
	MealType  string    `json:"meal_type"`
	FoodName  string    `json:"food_name"`
	Calories  int       `json:"calories"`
	ProteinG  *float64  `json:"protein_g,omitempty"`
	CarbsG    *float64  `json:"carbs_g,omitempty"`
	FatG      *float64  `json:"fat_g,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type calorieSummaryResponse struct {
	Date          string         `json:"date"`
	TotalCalories int            `json:"total_calories"`
	ByMealType    map[string]int `json:"by_meal_type"`
}
