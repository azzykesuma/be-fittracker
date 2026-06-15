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
	CreatedAt    time.Time
}

func (repo *Repository) CreateUser(ctx context.Context, id, name, email, passwordHash string) (userRecord, error) {
	var user userRecord
	err := repo.db.QueryRow(ctx, `
		INSERT INTO users (id, name, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, password_hash, created_at
	`, id, name, email, passwordHash).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	return user, err
}

func (repo *Repository) FindUserByEmail(ctx context.Context, email string) (userRecord, error) {
	var user userRecord
	err := repo.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	return user, err
}

func (repo *Repository) FindUserByID(ctx context.Context, id string) (userRecord, error) {
	var user userRecord
	err := repo.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
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
