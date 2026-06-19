package progress

import "testing"

func TestParseDateRangeDefaultsToWeek(t *testing.T) {
	from, to, err := parseDateRange("", "")
	if err != nil {
		t.Fatalf("parseDateRange returned error: %v", err)
	}

	if got := int(to.Sub(from).Hours() / 24); got != 6 {
		t.Fatalf("expected inclusive 7 day range, got %d days apart", got)
	}
}

func TestParseDateRangeWithOnlyToUsesPreviousWeek(t *testing.T) {
	from, to, err := parseDateRange("", "2026-06-15")
	if err != nil {
		t.Fatalf("parseDateRange returned error: %v", err)
	}

	if to.Format("2006-01-02") != "2026-06-15" {
		t.Fatalf("unexpected to date: %s", to.Format("2006-01-02"))
	}
	if from.Format("2006-01-02") != "2026-06-09" {
		t.Fatalf("unexpected from date: %s", from.Format("2006-01-02"))
	}
}

func TestParseDateRangeWithOnlyFromUsesFollowingWeek(t *testing.T) {
	from, to, err := parseDateRange("2026-06-15", "")
	if err != nil {
		t.Fatalf("parseDateRange returned error: %v", err)
	}

	if from.Format("2006-01-02") != "2026-06-15" {
		t.Fatalf("unexpected from date: %s", from.Format("2006-01-02"))
	}
	if to.Format("2006-01-02") != "2026-06-21" {
		t.Fatalf("unexpected to date: %s", to.Format("2006-01-02"))
	}
}

func TestCalculateBMI(t *testing.T) {
	heightCM := 175
	bmi := calculateBMI(78.5, &heightCM)
	if bmi == nil {
		t.Fatal("expected bmi")
	}
	if *bmi != 25.63 {
		t.Fatalf("expected bmi 25.63, got %.2f", *bmi)
	}
}

func TestCalculateBMIReturnsNilWithoutHeight(t *testing.T) {
	if bmi := calculateBMI(78.5, nil); bmi != nil {
		t.Fatalf("expected nil bmi, got %.2f", *bmi)
	}
}

func TestEstimateBodyFatPercentageMale(t *testing.T) {
	heightCM := 175
	neck := 38.0
	waist := 84.0
	bodyFat := estimateBodyFatPercentage(createBodyMeasurementRequest{
		NeckCM:  &neck,
		WaistCM: &waist,
	}, &heightCM, "male")
	if bodyFat == nil {
		t.Fatal("expected body fat estimate")
	}
	if *bodyFat != 16.15 {
		t.Fatalf("expected body fat 16.15, got %.2f", *bodyFat)
	}
}

func TestEstimateBodyFatPercentageFemale(t *testing.T) {
	heightCM := 175
	neck := 38.0
	waist := 84.0
	hips := 96.0
	bodyFat := estimateBodyFatPercentage(createBodyMeasurementRequest{
		NeckCM:  &neck,
		WaistCM: &waist,
		HipsCM:  &hips,
	}, &heightCM, "female")
	if bodyFat == nil {
		t.Fatal("expected body fat estimate")
	}
	if *bodyFat != 26.83 {
		t.Fatalf("expected body fat 26.83, got %.2f", *bodyFat)
	}
}

func TestEstimateBodyFatPercentageReturnsNilWithoutHeight(t *testing.T) {
	neck := 38.0
	waist := 84.0
	bodyFat := estimateBodyFatPercentage(createBodyMeasurementRequest{
		NeckCM:  &neck,
		WaistCM: &waist,
	}, nil, "male")
	if bodyFat != nil {
		t.Fatalf("expected nil body fat without height, got %.2f", *bodyFat)
	}
}
