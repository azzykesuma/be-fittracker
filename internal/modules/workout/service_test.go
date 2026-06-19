package workout

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"be-fittracker/internal/database"
)

type mockQuerier struct {
	database.Querier
	queryRowFunc func(sql string, args ...any) pgx.Row
}

func (m *mockQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(sql, args...)
	}
	return &mockRow{err: pgx.ErrNoRows}
}

type mockRow struct {
	err      error
	scanFunc func(dest ...any) error
}

func (r *mockRow) Scan(dest ...any) error {
	if r.scanFunc != nil {
		return r.scanFunc(dest...)
	}
	return r.err
}

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

func TestListSessionsRejectsInvalidStatus(t *testing.T) {
	svc := NewService(NewRepository(&mockQuerier{}))
	_, err := svc.ListSessions(context.Background(), "user-id", listSessionsFilter{Status: "invalid-status"})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestStartSessionRejectsEmptyPlanID(t *testing.T) {
	svc := NewService(NewRepository(&mockQuerier{}))
	_, err := svc.StartSession(context.Background(), "user-id", startSessionRequest{WorkoutPlanID: ""})
	if err == nil {
		t.Fatal("expected error for empty plan ID")
	}
}

func TestLogSetValidation(t *testing.T) {
	t.Run("session not found", func(t *testing.T) {
		mq := &mockQuerier{
			queryRowFunc: func(sql string, args ...any) pgx.Row {
				return &mockRow{err: pgx.ErrNoRows}
			},
		}
		svc := NewService(NewRepository(mq))
		_, err := svc.LogSet(context.Background(), "session-id", "user-id", logSetRequest{
			ExerciseName: "Bench Press",
			SetNumber:    1,
			Reps:         10,
		})
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected pgx.ErrNoRows, got %v", err)
		}
	})

	t.Run("empty exercise name and ID", func(t *testing.T) {
		mq := &mockQuerier{
			queryRowFunc: func(sql string, args ...any) pgx.Row {
				// Mock FindSession success by scanning successfully
				return &mockRow{scanFunc: func(dest ...any) error {
					return nil
				}}
			},
		}
		svc := NewService(NewRepository(mq))
		_, err := svc.LogSet(context.Background(), "session-id", "user-id", logSetRequest{
			SetNumber: 1,
			Reps:      10,
		})
		if err == nil || err.Error() != "exercise_name or exercise_id is required" {
			t.Fatalf("expected empty exercise name/id error, got %v", err)
		}
	})

	t.Run("invalid set number", func(t *testing.T) {
		mq := &mockQuerier{
			queryRowFunc: func(sql string, args ...any) pgx.Row {
				return &mockRow{scanFunc: func(dest ...any) error {
					return nil
				}}
			},
		}
		svc := NewService(NewRepository(mq))
		_, err := svc.LogSet(context.Background(), "session-id", "user-id", logSetRequest{
			ExerciseName: "Bench Press",
			SetNumber:    0,
			Reps:         10,
		})
		if err == nil || err.Error() != "set_number must be greater than 0" {
			t.Fatalf("expected set_number error, got %v", err)
		}
	})

	t.Run("negative reps", func(t *testing.T) {
		mq := &mockQuerier{
			queryRowFunc: func(sql string, args ...any) pgx.Row {
				return &mockRow{scanFunc: func(dest ...any) error {
					return nil
				}}
			},
		}
		svc := NewService(NewRepository(mq))
		_, err := svc.LogSet(context.Background(), "session-id", "user-id", logSetRequest{
			ExerciseName: "Bench Press",
			SetNumber:    1,
			Reps:         -5,
		})
		if err == nil || err.Error() != "reps must be non-negative" {
			t.Fatalf("expected reps error, got %v", err)
		}
	})

	t.Run("negative weight", func(t *testing.T) {
		mq := &mockQuerier{
			queryRowFunc: func(sql string, args ...any) pgx.Row {
				return &mockRow{scanFunc: func(dest ...any) error {
					return nil
				}}
			},
		}
		svc := NewService(NewRepository(mq))
		weight := -10.5
		_, err := svc.LogSet(context.Background(), "session-id", "user-id", logSetRequest{
			ExerciseName: "Bench Press",
			SetNumber:    1,
			Reps:         10,
			WeightKG:     &weight,
		})
		if err == nil || err.Error() != "weight_kg must be non-negative" {
			t.Fatalf("expected weight error, got %v", err)
		}
	})
}
