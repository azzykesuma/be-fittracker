package progress

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"be-fittracker/internal/database"
)

type Repository struct {
	db database.Querier
}

type bodyMeasurementRecord struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type userBodyProfile struct {
	HeightCM *int
	Gender   string
}

func NewRepository(db database.Querier) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) UserBodyProfile(ctx context.Context, userID string) (userBodyProfile, error) {
	var height sql.NullInt32
	var gender string
	err := repo.db.QueryRow(ctx, `
		SELECT height_cm, gender
		FROM users
		WHERE id = $1
	`, userID).Scan(&height, &gender)
	if err != nil {
		return userBodyProfile{}, err
	}

	var heightCM *int
	if height.Valid {
		heightVal := int(height.Int32)
		heightCM = &heightVal
	}
	return userBodyProfile{HeightCM: heightCM, Gender: gender}, nil
}

func (repo *Repository) LatestBodyMeasurementDate(ctx context.Context, userID string) (time.Time, error) {
	var latest time.Time
	err := repo.db.QueryRow(ctx, `
		SELECT log_date
		FROM body_measurement_logs
		WHERE user_id = $1
		ORDER BY log_date DESC
		LIMIT 1
	`, userID).Scan(&latest)
	return latest, err
}

func (repo *Repository) CreateBodyMeasurement(ctx context.Context, id string, userID string, req createBodyMeasurementRequest, logDate time.Time) (bodyMeasurementRecord, error) {
	var record bodyMeasurementRecord
	err := repo.db.QueryRow(ctx, `
		INSERT INTO body_measurement_logs (
			id,
			user_id,
			weight_kg,
			bmi,
			body_fat_percentage,
			neck_cm,
			shoulder_cm,
			chest_cm,
			waist_cm,
			belly_cm,
			hips_cm,
			left_bicep_cm,
			right_bicep_cm,
			left_forearm_cm,
			right_forearm_cm,
			left_thigh_cm,
			right_thigh_cm,
			left_calf_cm,
			right_calf_cm,
			notes,
			log_date
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NULLIF($20, ''), $21)
		RETURNING id, created_at, updated_at
	`, id, userID, req.WeightKG, req.BMI, req.BodyFatPercentage, req.NeckCM, req.ShoulderCM, req.ChestCM, req.WaistCM, req.BellyCM, req.HipsCM, req.LeftBicepCM, req.RightBicepCM, req.LeftForearmCM, req.RightForearmCM, req.LeftThighCM, req.RightThighCM, req.LeftCalfCM, req.RightCalfCM, req.Notes, logDate).Scan(&record.ID, &record.CreatedAt, &record.UpdatedAt)
	return record, err
}

func (repo *Repository) BodyMeasurementPoints(ctx context.Context, userID string, from time.Time, to time.Time) ([]bodyMeasurementPoint, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT log_date, weight_kg::float8, bmi::float8, body_fat_percentage::float8, waist_cm::float8
		FROM body_measurement_logs
		WHERE user_id = $1 AND log_date BETWEEN $2 AND $3
		ORDER BY log_date ASC
	`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []bodyMeasurementPoint{}
	for rows.Next() {
		var logDate time.Time
		var weight pgtype.Float8
		var bmi pgtype.Float8
		var bodyFat pgtype.Float8
		var waist pgtype.Float8
		if err := rows.Scan(&logDate, &weight, &bmi, &bodyFat, &waist); err != nil {
			return nil, err
		}

		items = append(items, bodyMeasurementPoint{
			Date:              logDate.Format("2006-01-02"),
			WeightKG:          nullableFloat64(weight),
			BMI:               nullableFloat64(bmi),
			BodyFatPercentage: nullableFloat64(bodyFat),
			WaistCM:           nullableFloat64(waist),
		})
	}

	return items, rows.Err()
}

func nullableFloat64(value pgtype.Float8) *float64 {
	if !value.Valid {
		return nil
	}
	return &value.Float64
}
