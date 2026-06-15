package habit

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

type habitRecord struct {
	ID             string
	Name           string
	Description    string
	Frequency      string
	TargetCount    int
	IsActive       bool
	CompletedToday bool
	CurrentStreak  int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (repo *Repository) List(ctx context.Context, userID string) ([]habitRecord, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT
			h.id,
			h.name,
			COALESCE(h.description, ''),
			h.frequency,
			h.target_count,
			h.is_active,
			EXISTS (
				SELECT 1 FROM habit_logs hl
				WHERE hl.habit_id = h.id AND hl.log_date = CURRENT_DATE AND hl.is_completed = TRUE
			) AS completed_today,
			h.created_at,
			h.updated_at
		FROM habits h
		WHERE h.user_id = $1 AND h.is_active = TRUE
		ORDER BY h.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}

	habits := []habitRecord{}
	for rows.Next() {
		var item habitRecord
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Frequency, &item.TargetCount, &item.IsActive, &item.CompletedToday, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		habits = append(habits, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	for i := range habits {
		habits[i].CurrentStreak = repo.currentStreak(ctx, habits[i].ID)
	}

	return habits, nil
}

func (repo *Repository) Create(ctx context.Context, id, userID, name, description, frequency string, targetCount int) (habitRecord, error) {
	var item habitRecord
	err := repo.db.QueryRow(ctx, `
		INSERT INTO habits (id, user_id, name, description, frequency, target_count)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6)
		RETURNING id, name, COALESCE(description, ''), frequency, target_count, is_active, created_at, updated_at
	`, id, userID, name, description, frequency, targetCount).Scan(&item.ID, &item.Name, &item.Description, &item.Frequency, &item.TargetCount, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (repo *Repository) CompleteToday(ctx context.Context, id, habitID, userID string) error {
	_, err := repo.db.Exec(ctx, `
		INSERT INTO habit_logs (id, habit_id, user_id, log_date, completed_count, is_completed)
		VALUES ($1, $2, $3, CURRENT_DATE, 1, TRUE)
		ON CONFLICT (habit_id, log_date)
		DO UPDATE SET completed_count = habit_logs.completed_count + 1, is_completed = TRUE
	`, id, habitID, userID)
	return err
}

func (repo *Repository) UncompleteToday(ctx context.Context, habitID, userID string) error {
	_, err := repo.db.Exec(ctx, `
		DELETE FROM habit_logs
		WHERE habit_id = $1 AND user_id = $2 AND log_date = CURRENT_DATE
	`, habitID, userID)
	return err
}

func (repo *Repository) BelongsToUser(ctx context.Context, habitID, userID string) (bool, error) {
	var exists bool
	err := repo.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM habits WHERE id = $1 AND user_id = $2)`, habitID, userID).Scan(&exists)
	return exists, err
}

func (repo *Repository) currentStreak(ctx context.Context, habitID string) int {
	rows, err := repo.db.Query(ctx, `
		SELECT log_date
		FROM habit_logs
		WHERE habit_id = $1 AND is_completed = TRUE
		ORDER BY log_date DESC
		LIMIT 60
	`, habitID)
	if err != nil {
		return 0
	}
	defer rows.Close()

	streak := 0
	expected := time.Now().Truncate(24 * time.Hour)
	for rows.Next() {
		var logDate time.Time
		if err := rows.Scan(&logDate); err != nil {
			return streak
		}
		if !sameDate(logDate, expected) {
			break
		}
		streak++
		expected = expected.AddDate(0, 0, -1)
	}
	return streak
}

func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
