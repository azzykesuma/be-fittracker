package auth

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

type userRecord struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	FitnessGoal  *string
	HeightCM     *int
	WeightKG     *float64
	Gender       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (repo *Repository) CreateUser(ctx context.Context, id, name, email, passwordHash, gender string) (userRecord, error) {
	var user userRecord
	err := repo.db.QueryRow(ctx, `
		INSERT INTO users (id, name, email, password_hash, gender)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, email, password_hash, fitness_goal, height_cm, weight_kg, gender, created_at, updated_at
	`, id, name, email, passwordHash, gender).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.FitnessGoal, &user.HeightCM, &user.WeightKG, &user.Gender, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (repo *Repository) FindUserByEmail(ctx context.Context, email string) (userRecord, error) {
	var user userRecord
	err := repo.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, fitness_goal, height_cm, weight_kg, gender, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.FitnessGoal, &user.HeightCM, &user.WeightKG, &user.Gender, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (repo *Repository) FindUserByID(ctx context.Context, id string) (userRecord, error) {
	var user userRecord
	err := repo.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, fitness_goal, height_cm, weight_kg, gender, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.FitnessGoal, &user.HeightCM, &user.WeightKG, &user.Gender, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (repo *Repository) UpdateUser(ctx context.Context, id string, name string, fitnessGoal *string, heightCM *int, weightKG *float64, gender string) (userRecord, error) {
	var user userRecord
	err := repo.db.QueryRow(ctx, `
		UPDATE users
		SET name = $2, fitness_goal = $3, height_cm = $4, weight_kg = $5, gender = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, email, password_hash, fitness_goal, height_cm, weight_kg, gender, created_at, updated_at
	`, id, name, fitnessGoal, heightCM, weightKG, gender).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.FitnessGoal, &user.HeightCM, &user.WeightKG, &user.Gender, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (repo *Repository) CreateRefreshSession(ctx context.Context, id, userID, tokenHash string, expiresAt time.Time) error {
	_, err := repo.db.Exec(ctx, `
		INSERT INTO refresh_sessions (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`, id, userID, tokenHash, expiresAt)
	return err
}

func (repo *Repository) FindRefreshSession(ctx context.Context, tokenHash string) (string, time.Time, bool, error) {
	var userID string
	var expiresAt time.Time
	var revokedAt *time.Time
	err := repo.db.QueryRow(ctx, `
		SELECT user_id, expires_at, revoked_at
		FROM refresh_sessions
		WHERE token_hash = $1
	`, tokenHash).Scan(&userID, &expiresAt, &revokedAt)
	return userID, expiresAt, revokedAt != nil, err
}

func (repo *Repository) RevokeRefreshSession(ctx context.Context, tokenHash string) error {
	_, err := repo.db.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = NOW(), rotated_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	return err
}
