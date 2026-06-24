package progress

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo  *Repository
	redis *redis.Client
}

func NewService(repo *Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, redis: rdb}
}

type bodyMeasurementQuery struct {
	From string
	To   string
}

type bodyMeasurementPoint struct {
	Date              string   `json:"date"`
	WeightKG          *float64 `json:"weight_kg"`
	BMI               *float64 `json:"bmi"`
	BodyFatPercentage *float64 `json:"body_fat_percentage"`
	WaistCM           *float64 `json:"waist_cm"`
	ImageURL          *string  `json:"image_url,omitempty"`
}

type progressPhotoPoint struct {
	Date              string   `json:"date"`
	ImageURL          string   `json:"image_url"`
	WeightKG          *float64 `json:"weight_kg,omitempty"`
	BodyFatPercentage *float64 `json:"body_fat_percentage,omitempty"`
}

type createBodyMeasurementRequest struct {
	BMI               *float64 `json:"-"`
	BodyFatPercentage *float64 `json:"-"`
	WeightKG          float64  `json:"weight_kg"`
	NeckCM            *float64 `json:"neck_cm"`
	ShoulderCM        *float64 `json:"shoulder_cm"`
	ChestCM           *float64 `json:"chest_cm"`
	WaistCM           *float64 `json:"waist_cm"`
	BellyCM           *float64 `json:"belly_cm"`
	HipsCM            *float64 `json:"hips_cm"`
	LeftBicepCM       *float64 `json:"left_bicep_cm"`
	RightBicepCM      *float64 `json:"right_bicep_cm"`
	LeftForearmCM     *float64 `json:"left_forearm_cm"`
	RightForearmCM    *float64 `json:"right_forearm_cm"`
	LeftThighCM       *float64 `json:"left_thigh_cm"`
	RightThighCM      *float64 `json:"right_thigh_cm"`
	LeftCalfCM        *float64 `json:"left_calf_cm"`
	RightCalfCM       *float64 `json:"right_calf_cm"`
	Notes             string   `json:"notes"`
	LogDate           string   `json:"log_date"`
	ImageURL          string   `json:"image_url"`
}

type bodyMeasurementResponse struct {
	ID                string   `json:"id"`
	WeightKG          float64  `json:"weight_kg"`
	BMI               *float64 `json:"bmi,omitempty"`
	BodyFatPercentage *float64 `json:"body_fat_percentage,omitempty"`
	NeckCM            *float64 `json:"neck_cm,omitempty"`
	ShoulderCM        *float64 `json:"shoulder_cm,omitempty"`
	ChestCM           *float64 `json:"chest_cm,omitempty"`
	WaistCM           *float64 `json:"waist_cm,omitempty"`
	BellyCM           *float64 `json:"belly_cm,omitempty"`
	HipsCM            *float64 `json:"hips_cm,omitempty"`
	LeftBicepCM       *float64 `json:"left_bicep_cm,omitempty"`
	RightBicepCM      *float64 `json:"right_bicep_cm,omitempty"`
	LeftForearmCM     *float64 `json:"left_forearm_cm,omitempty"`
	RightForearmCM    *float64 `json:"right_forearm_cm,omitempty"`
	LeftThighCM       *float64 `json:"left_thigh_cm,omitempty"`
	RightThighCM      *float64 `json:"right_thigh_cm,omitempty"`
	LeftCalfCM        *float64 `json:"left_calf_cm,omitempty"`
	RightCalfCM       *float64 `json:"right_calf_cm,omitempty"`
	Notes             string   `json:"notes,omitempty"`
	LogDate           string   `json:"log_date"`
	ImageURL          string   `json:"image_url,omitempty"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

func (svc *Service) BodyMeasurements(ctx context.Context, userID string, query bodyMeasurementQuery) ([]bodyMeasurementPoint, error) {
	from, to, err := svc.parseBodyMeasurementDateRange(ctx, userID, query.From, query.To)
	if err != nil {
		return nil, err
	}
	if to.Before(from) {
		return nil, errors.New("to must be on or after from")
	}

	cacheKey := fmt.Sprintf("fitflow:progress:measurements:%s:%s:%s", userID, from.Format("2006-01-02"), to.Format("2006-01-02"))

	if svc.redis != nil {
		if val, err := svc.redis.Get(ctx, cacheKey).Result(); err == nil {
			var cached []bodyMeasurementPoint
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				return cached, nil
			}
		}
	}

	points, err := svc.repo.BodyMeasurementPoints(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}

	if svc.redis != nil {
		if val, err := json.Marshal(points); err == nil {
			_ = svc.redis.Set(ctx, cacheKey, val, 10*time.Minute).Err()
		}
	}

	return points, nil
}

func (svc *Service) ProgressPhotos(ctx context.Context, userID string) ([]progressPhotoPoint, error) {
	return svc.repo.ProgressPhotos(ctx, userID)
}

func (svc *Service) CreateBodyMeasurement(ctx context.Context, userID string, req createBodyMeasurementRequest) (bodyMeasurementResponse, error) {
	if req.WeightKG <= 0 {
		return bodyMeasurementResponse{}, errors.New("weight_kg must be greater than 0")
	}

	logDate, err := parseDate(req.LogDate)
	if err != nil {
		return bodyMeasurementResponse{}, errors.New("log_date must be YYYY-MM-DD")
	}

	profile, err := svc.repo.UserBodyProfile(ctx, userID)
	if err != nil {
		return bodyMeasurementResponse{}, err
	}
	req.BMI = calculateBMI(req.WeightKG, profile.HeightCM)
	req.BodyFatPercentage = estimateBodyFatPercentage(req, profile.HeightCM, profile.Gender)

	record, err := svc.repo.CreateBodyMeasurement(ctx, uuid.NewString(), userID, req, logDate)
	if err != nil {
		return bodyMeasurementResponse{}, err
	}

	if svc.redis != nil {
		pattern := fmt.Sprintf("fitflow:progress:measurements:%s:*", userID)
		if keys, err := svc.redis.Keys(ctx, pattern).Result(); err == nil && len(keys) > 0 {
			_ = svc.redis.Del(ctx, keys...).Err()
		}
	}

	req.LogDate = logDate.Format("2006-01-02")
	return bodyMeasurementResponse{
		ID:                record.ID,
		WeightKG:          req.WeightKG,
		BMI:               req.BMI,
		BodyFatPercentage: req.BodyFatPercentage,
		NeckCM:            req.NeckCM,
		ShoulderCM:        req.ShoulderCM,
		ChestCM:           req.ChestCM,
		WaistCM:           req.WaistCM,
		BellyCM:           req.BellyCM,
		HipsCM:            req.HipsCM,
		LeftBicepCM:       req.LeftBicepCM,
		RightBicepCM:      req.RightBicepCM,
		LeftForearmCM:     req.LeftForearmCM,
		RightForearmCM:    req.RightForearmCM,
		LeftThighCM:       req.LeftThighCM,
		RightThighCM:      req.RightThighCM,
		LeftCalfCM:        req.LeftCalfCM,
		RightCalfCM:       req.RightCalfCM,
		Notes:             req.Notes,
		LogDate:           req.LogDate,
		ImageURL:          req.ImageURL,
		CreatedAt:         record.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         record.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func calculateBMI(weightKG float64, heightCM *int) *float64 {
	if heightCM == nil || *heightCM <= 0 || weightKG <= 0 {
		return nil
	}

	heightM := float64(*heightCM) / 100
	return floatPtr(round2(weightKG / (heightM * heightM)))
}

func estimateBodyFatPercentage(req createBodyMeasurementRequest, heightCM *int, gender string) *float64 {
	if heightCM == nil || *heightCM <= 0 || req.NeckCM == nil || *req.NeckCM <= 0 {
		return nil
	}

	waist := req.WaistCM
	if waist == nil {
		waist = req.BellyCM
	}
	if waist == nil || *waist <= 0 {
		return nil
	}

	height := float64(*heightCM)

	if gender == "female" {
		if req.HipsCM == nil || *req.HipsCM <= 0 {
			return nil
		}
		term := *waist + *req.HipsCM - *req.NeckCM
		if term <= 0 {
			return nil
		}
		value := 495/(1.29579-0.35004*math.Log10(term)+0.22100*math.Log10(height)) - 450
		return boundedPercentage(value)
	}

	// Default to male
	term := *waist - *req.NeckCM
	if term <= 0 {
		return nil
	}
	value := 495/(1.0324-0.19077*math.Log10(term)+0.15456*math.Log10(height)) - 450
	return boundedPercentage(value)
}

func boundedPercentage(value float64) *float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 100 {
		return nil
	}
	return floatPtr(round2(value))
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func floatPtr(value float64) *float64 {
	return &value
}

func (svc *Service) parseBodyMeasurementDateRange(ctx context.Context, userID string, fromValue string, toValue string) (time.Time, time.Time, error) {
	if strings.TrimSpace(fromValue) != "" || strings.TrimSpace(toValue) != "" {
		return parseDateRange(fromValue, toValue)
	}

	latest, err := svc.repo.LatestBodyMeasurementDate(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return parseDateRange("", "")
		}
		return time.Time{}, time.Time{}, err
	}

	return latest.AddDate(0, 0, -6), latest, nil
}

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}

func parseDateRange(fromValue string, toValue string) (time.Time, time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	fromValue = strings.TrimSpace(fromValue)
	toValue = strings.TrimSpace(toValue)
	if fromValue == "" && toValue == "" {
		return today.AddDate(0, 0, -6), today, nil
	}

	if fromValue == "" {
		to, err := parseDate(toValue)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("to must be YYYY-MM-DD")
		}
		return to.AddDate(0, 0, -6), to, nil
	}

	from, err := parseDate(fromValue)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("from must be YYYY-MM-DD")
	}
	if toValue == "" {
		return from, from.AddDate(0, 0, 6), nil
	}

	to, err := parseDate(toValue)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("to must be YYYY-MM-DD")
	}
	return from, to, nil
}
