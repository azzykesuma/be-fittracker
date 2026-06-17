package meal

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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

	records, err := svc.repo.List(ctx, userID, from, to, mealType)
	if err != nil {
		return nil, err
	}
	items := make([]mealLogResponse, 0, len(records))
	for _, record := range records {
		items = append(items, toMealLogResponse(record))
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
	return toMealLogResponse(record), nil
}

func (svc *Service) Delete(ctx context.Context, id string, userID string) error {
	return svc.repo.Delete(ctx, id, userID)
}

func (svc *Service) CalorieSummary(ctx context.Context, userID string, dateValue string) (calorieSummaryResponse, error) {
	date, err := parseDateOrToday(dateValue)
	if err != nil {
		return calorieSummaryResponse{}, err
	}
	byMealType, err := svc.repo.CalorieSummary(ctx, userID, date)
	if err != nil {
		return calorieSummaryResponse{}, err
	}
	total := 0
	for _, calories := range byMealType {
		total += calories
	}
	return calorieSummaryResponse{Date: date.Format("2006-01-02"), TotalCalories: total, ByMealType: byMealType}, nil
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
