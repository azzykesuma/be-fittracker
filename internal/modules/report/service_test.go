package report

import (
	"bytes"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestBuildExcelWorkbook(t *testing.T) {
	svc := &Service{}

	now := time.Now()
	height := 180
	weight := 82.5
	bmi := 25.46
	bodyFat := 18.2

	summary := UserSummaryRecord{
		Name:             "John Doe",
		Email:            "john@example.com",
		HeightCM:         &height,
		WeightKG:         &weight,
		BMI:              &bmi,
		BodyFat:          &bodyFat,
		TotalWorkouts:    5,
		AvgDailyCalories: 2450.0,
		CreatedAt:        now.AddDate(-1, 0, 0),
	}

	progressLogs := []ProgressRecord{
		{
			LogDate:           now.AddDate(0, 0, -10),
			WeightKG:          83.0,
			BMI:               25.62,
			BodyFatPercentage: 18.5,
			NeckCM:            39.0,
			WaistCM:           88.0,
			Notes:             "Initial",
		},
		{
			LogDate:           now.AddDate(0, 0, -5),
			WeightKG:          82.5,
			BMI:               25.46,
			BodyFatPercentage: 18.2,
			NeckCM:            39.0,
			WaistCM:           87.5,
			Notes:             "Improved",
		},
	}

	mealLogs := []MealRecord{
		{
			MealDate: now.AddDate(0, 0, -1),
			MealType: "breakfast",
			FoodName: "Oatmeal",
			Calories: 450,
			ProteinG: 15.0,
			CarbsG:   60.0,
			FatG:     8.0,
			Notes:    "Post-workout",
		},
	}

	workoutLogs := []WorkoutRecord{
		{
			StartedAt:    now.AddDate(0, 0, -2),
			PlanName:     "Upper Body",
			Status:       "finished",
			SessionNotes: "Great session",
			ExerciseName: "Bench Press",
			SetNumber:    1,
			Reps:         8,
			WeightKG:     80.0,
			Completed:    true,
		},
	}

	data, fileName, err := svc.buildExcelWorkbook(summary, progressLogs, mealLogs, workoutLogs)
	if err != nil {
		t.Fatalf("failed to build Excel workbook: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("generated empty byte slice for Excel file")
	}

	if fileName == "" {
		t.Fatal("generated empty filename")
	}

	// Verify worksheets inside the output xlsx file
	reader := bytes.NewReader(data)
	f, err := excelize.OpenReader(reader)
	if err != nil {
		t.Fatalf("failed to read generated Excel file: %v", err)
	}
	defer f.Close()

	expectedSheets := []string{"Overview", "Progress Logs", "Meal Logs", "Workout Logs"}
	for _, expected := range expectedSheets {
		index, err := f.GetSheetIndex(expected)
		if err != nil || index == -1 {
			t.Fatalf("expected sheet %q not found or error occurred: %v", expected, err)
		}
	}
}
