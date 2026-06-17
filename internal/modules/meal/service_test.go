package meal

import "testing"

func TestValidateMealLogRequest(t *testing.T) {
	protein := 18.0
	req, mealDate, err := validateMealLogRequest(mealLogRequest{
		MealDate: "2026-06-15",
		MealType: "Breakfast",
		FoodName: " Oatmeal ",
		Calories: 420,
		ProteinG: &protein,
	})
	if err != nil {
		t.Fatalf("validateMealLogRequest returned error: %v", err)
	}
	if req.MealType != "breakfast" {
		t.Fatalf("expected normalized meal type, got %s", req.MealType)
	}
	if req.FoodName != "Oatmeal" {
		t.Fatalf("expected trimmed food name, got %s", req.FoodName)
	}
	if mealDate.Format("2006-01-02") != "2026-06-15" {
		t.Fatalf("unexpected meal date: %s", mealDate.Format("2006-01-02"))
	}
}

func TestValidateMealLogRequestRejectsInvalidMealType(t *testing.T) {
	_, _, err := validateMealLogRequest(mealLogRequest{MealDate: "2026-06-15", MealType: "brunch", FoodName: "Oatmeal"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseDateRangeDefaultsToWeek(t *testing.T) {
	from, to, err := parseDateRange("", "")
	if err != nil {
		t.Fatalf("parseDateRange returned error: %v", err)
	}
	if got := int(to.Sub(from).Hours() / 24); got != 6 {
		t.Fatalf("expected inclusive 7 day range, got %d days apart", got)
	}
}
