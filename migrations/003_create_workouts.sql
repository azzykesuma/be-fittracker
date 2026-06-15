CREATE TABLE workout_plans (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    scheduled_day VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workout_plans_user_id ON workout_plans(user_id);
CREATE INDEX idx_workout_plans_user_day ON workout_plans(user_id, scheduled_day);

CREATE TABLE exercises (
    id UUID PRIMARY KEY,
    workout_plan_id UUID NOT NULL REFERENCES workout_plans(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    muscle_group VARCHAR(50),
    target_sets INT,
    target_reps INT,
    target_weight_kg DECIMAL(5,2),
    rest_seconds INT NOT NULL DEFAULT 60,
    order_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exercises_workout_plan_id ON exercises(workout_plan_id);

CREATE TABLE workout_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workout_plan_id UUID REFERENCES workout_plans(id) ON DELETE SET NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    status VARCHAR(30) NOT NULL DEFAULT 'in_progress',
    notes TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workout_sessions_user_started ON workout_sessions(user_id, started_at);
CREATE INDEX idx_workout_sessions_user_status ON workout_sessions(user_id, status);

CREATE TABLE workout_set_logs (
    id UUID PRIMARY KEY,
    workout_session_id UUID NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
    exercise_id UUID REFERENCES exercises(id) ON DELETE SET NULL,
    exercise_name VARCHAR(100) NOT NULL,
    set_number INT NOT NULL,
    reps INT NOT NULL,
    weight_kg DECIMAL(5,2),
    completed BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workout_set_logs_session_id ON workout_set_logs(workout_session_id);
CREATE INDEX idx_workout_set_logs_exercise_id ON workout_set_logs(exercise_id);
