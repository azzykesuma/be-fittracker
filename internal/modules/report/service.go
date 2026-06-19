package report

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (svc *Service) GenerateSummaryReport(ctx context.Context, userID string) ([]byte, string, error) {
	// 1. Fetch all required data
	summary, err := svc.repo.GetUserSummary(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch user summary: %w", err)
	}

	progressLogs, err := svc.repo.GetProgressLogs(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch progress logs: %w", err)
	}

	mealLogs, err := svc.repo.GetMealLogs(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch meal logs: %w", err)
	}

	workoutLogs, err := svc.repo.GetWorkoutLogs(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch workout logs: %w", err)
	}

	return svc.buildExcelWorkbook(summary, progressLogs, mealLogs, workoutLogs)
}

func (svc *Service) buildExcelWorkbook(summary UserSummaryRecord, progressLogs []ProgressRecord, mealLogs []MealRecord, workoutLogs []WorkoutRecord) ([]byte, string, error) {
	// 2. Initialize new excel file
	f := excelize.NewFile()
	defer f.Close()

	// Default sheet is usually "Sheet1", we will rename it to "Overview"
	f.SetSheetName("Sheet1", "Overview")

	// Create remaining sheets
	f.NewSheet("Progress Logs")
	f.NewSheet("Meal Logs")
	f.NewSheet("Workout Logs")

	// Create styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 18, Color: "1F4E78", Family: "Segoe UI"},
	})
	secHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 13, Color: "1F4E78", Family: "Segoe UI"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"D9E1F2"}, Pattern: 1},
	})
	labelStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Color: "595959", Family: "Segoe UI"},
	})
	valueStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10, Family: "Segoe UI"},
	})
	boldValueStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000", Family: "Segoe UI"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF", Family: "Segoe UI"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1F4E78"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	altRowStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F2F2F2"}, Pattern: 1},
		Font: &excelize.Font{Size: 10, Family: "Segoe UI"},
	})
	normalRowStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10, Family: "Segoe UI"},
	})

	// ----------------------------------------------------
	// A. Write Overview Sheet
	// ----------------------------------------------------
	sheet := "Overview"
	enableGridlines(f, sheet)

	f.SetCellValue(sheet, "A1", "FitFlow Personal Summary Report")
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	f.SetRowHeight(sheet, 1, 30)

	f.SetCellValue(sheet, "A2", fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")))
	f.SetCellStyle(sheet, "A2", "A2", valueStyle)

	// User Profile section
	f.SetCellValue(sheet, "A4", "USER PROFILE")
	f.MergeCell(sheet, "A4", "B4")
	f.SetCellStyle(sheet, "A4", "B4", secHeaderStyle)

	profileData := [][]any{
		{"Name:", summary.Name},
		{"Email:", summary.Email},
		{"Member Since:", summary.CreatedAt.Format("2006-01-02")},
		{"Latest Height (cm):", nullableVal(summary.HeightCM)},
		{"Latest Weight (kg):", nullableVal(summary.WeightKG)},
		{"Latest BMI:", nullableVal(summary.BMI)},
		{"Latest Body Fat (%):", nullableVal(summary.BodyFat)},
	}
	rowOffset := 5
	for _, data := range profileData {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowOffset), data[0])
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowOffset), fmt.Sprintf("A%d", rowOffset), labelStyle)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowOffset), data[1])
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", rowOffset), fmt.Sprintf("B%d", rowOffset), valueStyle)
		rowOffset++
	}

	// Performance Summary section
	f.SetCellValue(sheet, "D4", "ACTIVITY SUMMARY")
	f.MergeCell(sheet, "D4", "E4")
	f.SetCellStyle(sheet, "D4", "E4", secHeaderStyle)

	summaryData := [][]any{
		{"Workouts Logged:", summary.TotalWorkouts},
		{"Avg. Daily Calories:", fmt.Sprintf("%.0f kcal", summary.AvgDailyCalories)},
	}
	rowOffset = 5
	for _, data := range summaryData {
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowOffset), data[0])
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", rowOffset), fmt.Sprintf("D%d", rowOffset), labelStyle)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowOffset), data[1])
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", rowOffset), fmt.Sprintf("E%d", rowOffset), boldValueStyle)
		rowOffset++
	}
	autoFitColumns(f, sheet, 5)

	// ----------------------------------------------------
	// B. Write Progress Sheet
	// ----------------------------------------------------
	sheet = "Progress Logs"
	enableGridlines(f, sheet)

	progressHeaders := []string{
		"Date", "Weight (kg)", "BMI", "Body Fat %", "Neck (cm)", "Shoulder (cm)", "Chest (cm)",
		"Waist (cm)", "Belly (cm)", "Hips (cm)", "L Bicep (cm)", "R Bicep (cm)", "L Forearm (cm)",
		"R Forearm (cm)", "L Thigh (cm)", "R Thigh (cm)", "L Calf (cm)", "R Calf (cm)", "Notes",
	}
	for colIdx, header := range progressHeaders {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheet, 1, 24)

	for rowIdx, rec := range progressLogs {
		rowNum := rowIdx + 2
		rowStyle := normalRowStyle
		if rowNum%2 == 1 {
			rowStyle = altRowStyle
		}

		vals := []any{
			rec.LogDate.Format("2006-01-02"),
			zeroToNil(rec.WeightKG),
			zeroToNil(rec.BMI),
			zeroToNil(rec.BodyFatPercentage),
			zeroToNil(rec.NeckCM),
			zeroToNil(rec.ShoulderCM),
			zeroToNil(rec.ChestCM),
			zeroToNil(rec.WaistCM),
			zeroToNil(rec.BellyCM),
			zeroToNil(rec.HipsCM),
			zeroToNil(rec.LeftBicepCM),
			zeroToNil(rec.RightBicepCM),
			zeroToNil(rec.LeftForearmCM),
			zeroToNil(rec.RightForearmCM),
			zeroToNil(rec.LeftThighCM),
			zeroToNil(rec.RightThighCM),
			zeroToNil(rec.LeftCalfCM),
			zeroToNil(rec.RightCalfCM),
			rec.Notes,
		}
		for colIdx, val := range vals {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowNum)
			f.SetCellValue(sheet, cell, val)
			f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}
	autoFitColumns(f, sheet, len(progressHeaders))

	// ----------------------------------------------------
	// C. Write Meals Sheet
	// ----------------------------------------------------
	sheet = "Meal Logs"
	enableGridlines(f, sheet)

	mealHeaders := []string{"Date", "Meal Type", "Food Name", "Calories", "Protein (g)", "Carbs (g)", "Fat (g)", "Notes"}
	for colIdx, header := range mealHeaders {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheet, 1, 24)

	for rowIdx, rec := range mealLogs {
		rowNum := rowIdx + 2
		rowStyle := normalRowStyle
		if rowNum%2 == 1 {
			rowStyle = altRowStyle
		}

		vals := []any{
			rec.MealDate.Format("2006-01-02"),
			rec.MealType,
			rec.FoodName,
			rec.Calories,
			zeroToNil(rec.ProteinG),
			zeroToNil(rec.CarbsG),
			zeroToNil(rec.FatG),
			rec.Notes,
		}
		for colIdx, val := range vals {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowNum)
			f.SetCellValue(sheet, cell, val)
			f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}
	autoFitColumns(f, sheet, len(mealHeaders))

	// ----------------------------------------------------
	// D. Write Workouts Sheet
	// ----------------------------------------------------
	sheet = "Workout Logs"
	enableGridlines(f, sheet)

	workoutHeaders := []string{"Date/Time", "Plan Name", "Session Status", "Session Notes", "Exercise Name", "Set #", "Reps", "Weight (kg)", "Completed"}
	for colIdx, header := range workoutHeaders {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheet, 1, 24)

	for rowIdx, rec := range workoutLogs {
		rowNum := rowIdx + 2
		rowStyle := normalRowStyle
		if rowNum%2 == 1 {
			rowStyle = altRowStyle
		}

		compStr := "No"
		if rec.Completed {
			compStr = "Yes"
		}

		vals := []any{
			rec.StartedAt.Format("2006-01-02 15:04"),
			rec.PlanName,
			rec.Status,
			rec.SessionNotes,
			rec.ExerciseName,
			rec.SetNumber,
			rec.Reps,
			rec.WeightKG,
			compStr,
		}
		for colIdx, val := range vals {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowNum)
			f.SetCellValue(sheet, cell, val)
			f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}
	autoFitColumns(f, sheet, len(workoutHeaders))

	// Set active tab back to Overview
	f.SetActiveSheet(0)

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", fmt.Errorf("failed to write excel file to buffer: %w", err)
	}

	fileName := fmt.Sprintf("FitFlow_Summary_%s.xlsx", time.Now().Format("20060102"))
	return buf.Bytes(), fileName, nil
}

func enableGridlines(f *excelize.File, sheet string) {
	showGrid := true
	_ = f.SetSheetView(sheet, 0, &excelize.ViewOptions{
		ShowGridLines: &showGrid,
	})
}

func autoFitColumns(f *excelize.File, sheet string, maxCol int) {
	for colNum := 1; colNum <= maxCol; colNum++ {
		colName, _ := excelize.ColumnNumberToName(colNum)
		width := 14.0
		rows, err := f.GetRows(sheet)
		if err == nil {
			maxLen := 0
			for _, row := range rows {
				if colNum-1 < len(row) {
					l := len(row[colNum-1])
					if l > maxLen {
						maxLen = l
					}
				}
			}
			if maxLen > 0 {
				width = float64(maxLen) + 4.0
				if width < 12 {
					width = 12
				}
				if width > 40 {
					width = 40
				}
			}
		}
		_ = f.SetColWidth(sheet, colName, colName, width)
	}
}

func nullableVal[T any](ptr *T) any {
	if ptr == nil {
		return "N/A"
	}
	return *ptr
}

func zeroToNil(val float64) any {
	if val == 0.0 {
		return nil
	}
	return val
}
