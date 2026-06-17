package meal

import (
	"context"
	"time"

	"be-fittracker/internal/database"
)

type Repository struct {
	db database.Querier
}

func NewRepository(db database.Querier) *Repository {
	return &Repository{db: db}
}

type mealLogRecord struct {
	ID        string
	MealDate  time.Time
	MealType  string
	FoodName  string
	Calories  int
	ProteinG  *float64
	CarbsG    *float64
	FatG      *float64
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (repo *Repository) Create(ctx context.Context, id string, userID string, req mealLogRequest, mealDate time.Time) (mealLogRecord, error) {
	var record mealLogRecord
	err := repo.db.QueryRow(ctx, `
		INSERT INTO meal_logs (id, user_id, meal_date, meal_type, food_name, calories, protein_g, carbs_g, fat_g, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''))
		RETURNING id, meal_date, meal_type, food_name, calories, protein_g, carbs_g, fat_g, COALESCE(notes, ''), created_at, updated_at
	`, id, userID, mealDate, req.MealType, req.FoodName, req.Calories, req.ProteinG, req.CarbsG, req.FatG, req.Notes).Scan(&record.ID, &record.MealDate, &record.MealType, &record.FoodName, &record.Calories, &record.ProteinG, &record.CarbsG, &record.FatG, &record.Notes, &record.CreatedAt, &record.UpdatedAt)
	return record, err
}

func (repo *Repository) List(ctx context.Context, userID string, from time.Time, to time.Time, mealType string) ([]mealLogRecord, error) {
	query := `
		SELECT id, meal_date, meal_type, food_name, calories, protein_g, carbs_g, fat_g, COALESCE(notes, ''), created_at, updated_at
		FROM meal_logs
		WHERE user_id = $1 AND meal_date BETWEEN $2 AND $3
	`
	args := []any{userID, from, to}
	if mealType != "" {
		query += ` AND meal_type = $4`
		args = append(args, mealType)
	}
	query += ` ORDER BY meal_date DESC, created_at DESC`

	rows, err := repo.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []mealLogRecord{}
	for rows.Next() {
		var record mealLogRecord
		if err := rows.Scan(&record.ID, &record.MealDate, &record.MealType, &record.FoodName, &record.Calories, &record.ProteinG, &record.CarbsG, &record.FatG, &record.Notes, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, record)
	}
	return items, rows.Err()
}

func (repo *Repository) Find(ctx context.Context, id string, userID string) (mealLogRecord, error) {
	var record mealLogRecord
	err := repo.db.QueryRow(ctx, `
		SELECT id, meal_date, meal_type, food_name, calories, protein_g, carbs_g, fat_g, COALESCE(notes, ''), created_at, updated_at
		FROM meal_logs
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&record.ID, &record.MealDate, &record.MealType, &record.FoodName, &record.Calories, &record.ProteinG, &record.CarbsG, &record.FatG, &record.Notes, &record.CreatedAt, &record.UpdatedAt)
	return record, err
}

func (repo *Repository) Update(ctx context.Context, id string, userID string, req mealLogRequest, mealDate time.Time) (mealLogRecord, error) {
	var record mealLogRecord
	err := repo.db.QueryRow(ctx, `
		UPDATE meal_logs
		SET meal_date = $3,
		    meal_type = $4,
		    food_name = $5,
		    calories = $6,
		    protein_g = $7,
		    carbs_g = $8,
		    fat_g = $9,
		    notes = NULLIF($10, ''),
		    updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, meal_date, meal_type, food_name, calories, protein_g, carbs_g, fat_g, COALESCE(notes, ''), created_at, updated_at
	`, id, userID, mealDate, req.MealType, req.FoodName, req.Calories, req.ProteinG, req.CarbsG, req.FatG, req.Notes).Scan(&record.ID, &record.MealDate, &record.MealType, &record.FoodName, &record.Calories, &record.ProteinG, &record.CarbsG, &record.FatG, &record.Notes, &record.CreatedAt, &record.UpdatedAt)
	return record, err
}

func (repo *Repository) Delete(ctx context.Context, id string, userID string) error {
	_, err := repo.db.Exec(ctx, `DELETE FROM meal_logs WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (repo *Repository) CalorieSummary(ctx context.Context, userID string, date time.Time) (map[string]int, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT meal_type, COALESCE(SUM(calories), 0)::int
		FROM meal_logs
		WHERE user_id = $1 AND meal_date = $2
		GROUP BY meal_type
	`, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := map[string]int{"breakfast": 0, "lunch": 0, "dinner": 0, "snack": 0}
	for rows.Next() {
		var mealType string
		var calories int
		if err := rows.Scan(&mealType, &calories); err != nil {
			return nil, err
		}
		items[mealType] = calories
	}
	return items, rows.Err()
}
