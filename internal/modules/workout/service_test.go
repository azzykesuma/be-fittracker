package workout

import "testing"

func TestValidatePlanRequest(t *testing.T) {
	req, err := validatePlanRequest(workoutPlanRequest{Name: " Pull Day ", ScheduledDay: "Monday"})
	if err != nil {
		t.Fatalf("validatePlanRequest returned error: %v", err)
	}
	if req.Name != "Pull Day" {
		t.Fatalf("expected trimmed name, got %q", req.Name)
	}
	if req.ScheduledDay != "monday" {
		t.Fatalf("expected lowercase day, got %q", req.ScheduledDay)
	}
}

func TestValidatePlanRequestRejectsInvalidDay(t *testing.T) {
	_, err := validatePlanRequest(workoutPlanRequest{Name: "Pull Day", ScheduledDay: "funday"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateExerciseRequest(t *testing.T) {
	sets := 3
	reps := 10
	req, err := validateExerciseRequest(exerciseRequest{Name: " Curl ", TargetSets: &sets, TargetReps: &reps})
	if err != nil {
		t.Fatalf("validateExerciseRequest returned error: %v", err)
	}
	if req.Name != "Curl" {
		t.Fatalf("expected trimmed exercise name, got %q", req.Name)
	}
}
