package meal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo  *Repository
	redis *redis.Client
}

func NewService(repo *Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, redis: rdb}
}

func (svc *Service) Create(ctx context.Context, userID string, req mealLogRequest) (mealLogResponse, error) {
	req, mealDate, err := validateMealLogRequest(req)
	if err != nil {
		return mealLogResponse{}, err
	}
	record, err := svc.repo.Create(ctx, uuid.NewString(), userID, req, mealDate)
	if err != nil {
		return mealLogResponse{}, err
	}

	svc.invalidateCache(ctx, userID)

	return toMealLogResponse(record), nil
}

func (svc *Service) List(ctx context.Context, userID string, fromValue string, toValue string, mealType string) ([]mealLogResponse, error) {
	from, to, err := parseDateRange(fromValue, toValue)
	if err != nil {
		return nil, err
	}
	mealType = strings.ToLower(strings.TrimSpace(mealType))
	if mealType != "" && !validMealType(mealType) {
		return nil, errors.New("invalid meal_type")
	}

	cacheKey := fmt.Sprintf("fitflow:meal:list:%s:%s:%s:%s", userID, from.Format("2006-01-02"), to.Format("2006-01-02"), mealType)

	if svc.redis != nil {
		if val, err := svc.redis.Get(ctx, cacheKey).Result(); err == nil {
			var cached []mealLogResponse
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				return cached, nil
			}
		}
	}

	records, err := svc.repo.List(ctx, userID, from, to, mealType)
	if err != nil {
		return nil, err
	}
	items := make([]mealLogResponse, 0, len(records))
	for _, record := range records {
		items = append(items, toMealLogResponse(record))
	}

	if svc.redis != nil {
		if val, err := json.Marshal(items); err == nil {
			_ = svc.redis.Set(ctx, cacheKey, val, 10*time.Minute).Err()
		}
	}

	return items, nil
}

func (svc *Service) Find(ctx context.Context, id string, userID string) (mealLogResponse, error) {
	record, err := svc.repo.Find(ctx, id, userID)
	if err != nil {
		return mealLogResponse{}, err
	}
	return toMealLogResponse(record), nil
}

func (svc *Service) Update(ctx context.Context, id string, userID string, req mealLogRequest) (mealLogResponse, error) {
	req, mealDate, err := validateMealLogRequest(req)
	if err != nil {
		return mealLogResponse{}, err
	}
	record, err := svc.repo.Update(ctx, id, userID, req, mealDate)
	if err != nil {
		return mealLogResponse{}, err
	}

	svc.invalidateCache(ctx, userID)

	return toMealLogResponse(record), nil
}

func (svc *Service) Delete(ctx context.Context, id string, userID string) error {
	err := svc.repo.Delete(ctx, id, userID)
	if err == nil {
		svc.invalidateCache(ctx, userID)
	}
	return err
}

func (svc *Service) CalorieSummary(ctx context.Context, userID string, dateValue string) (calorieSummaryResponse, error) {
	date, err := parseDateOrToday(dateValue)
	if err != nil {
		return calorieSummaryResponse{}, err
	}

	cacheKey := fmt.Sprintf("fitflow:meal:calories:%s:%s", userID, date.Format("2006-01-02"))

	if svc.redis != nil {
		if val, err := svc.redis.Get(ctx, cacheKey).Result(); err == nil {
			var cached calorieSummaryResponse
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				return cached, nil
			}
		}
	}

	byMealType, err := svc.repo.CalorieSummary(ctx, userID, date)
	if err != nil {
		return calorieSummaryResponse{}, err
	}
	total := 0
	for _, calories := range byMealType {
		total += calories
	}
	res := calorieSummaryResponse{Date: date.Format("2006-01-02"), TotalCalories: total, ByMealType: byMealType}

	if svc.redis != nil {
		if val, err := json.Marshal(res); err == nil {
			_ = svc.redis.Set(ctx, cacheKey, val, 10*time.Minute).Err()
		}
	}

	return res, nil
}

func (svc *Service) invalidateCache(ctx context.Context, userID string) {
	if svc.redis == nil {
		return
	}
	patterns := []string{
		fmt.Sprintf("fitflow:meal:list:%s:*", userID),
		fmt.Sprintf("fitflow:meal:calories:%s:*", userID),
	}
	for _, pattern := range patterns {
		if keys, err := svc.redis.Keys(ctx, pattern).Result(); err == nil && len(keys) > 0 {
			_ = svc.redis.Del(ctx, keys...).Err()
		}
	}
}

func validateMealLogRequest(req mealLogRequest) (mealLogRequest, time.Time, error) {
	req.MealType = strings.ToLower(strings.TrimSpace(req.MealType))
	req.FoodName = strings.TrimSpace(req.FoodName)
	req.Notes = strings.TrimSpace(req.Notes)
	if req.FoodName == "" {
		return req, time.Time{}, errors.New("food_name is required")
	}
	if !validMealType(req.MealType) {
		return req, time.Time{}, errors.New("meal_type must be breakfast, lunch, dinner, or snack")
	}
	if req.Calories < 0 {
		return req, time.Time{}, errors.New("calories must be non-negative")
	}
	if negative(req.ProteinG) || negative(req.CarbsG) || negative(req.FatG) {
		return req, time.Time{}, errors.New("macro values must be non-negative")
	}
	mealDate, err := parseDate(req.MealDate)
	if err != nil {
		return req, time.Time{}, errors.New("meal_date must be YYYY-MM-DD")
	}
	return req, mealDate, nil
}

func validMealType(mealType string) bool {
	switch mealType {
	case "breakfast", "lunch", "dinner", "snack":
		return true
	default:
		return false
	}
}

func negative(value *float64) bool {
	return value != nil && *value < 0
}

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}

func parseDateOrToday(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value != "" {
		return parseDate(value)
	}
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
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
	if to.Before(from) {
		return time.Time{}, time.Time{}, errors.New("to must be on or after from")
	}
	return from, to, nil
}

func toMealLogResponse(record mealLogRecord) mealLogResponse {
	return mealLogResponse{
		ID:        record.ID,
		MealDate:  record.MealDate.Format("2006-01-02"),
		MealType:  record.MealType,
		FoodName:  record.FoodName,
		Calories:  record.Calories,
		ProteinG:  record.ProteinG,
		CarbsG:    record.CarbsG,
		FatG:      record.FatG,
		Notes:     record.Notes,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}
