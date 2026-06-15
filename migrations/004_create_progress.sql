CREATE TABLE IF NOT EXISTS body_measurement_logs (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  weight_kg DECIMAL(5,2) NOT NULL,
  bmi DECIMAL(5,2),
  body_fat_percentage DECIMAL(5,2),
  neck_cm DECIMAL(5,2),
  shoulder_cm DECIMAL(5,2),
  chest_cm DECIMAL(5,2),
  waist_cm DECIMAL(5,2),
  belly_cm DECIMAL(5,2),
  hips_cm DECIMAL(5,2),
  left_bicep_cm DECIMAL(5,2),
  right_bicep_cm DECIMAL(5,2),
  left_forearm_cm DECIMAL(5,2),
  right_forearm_cm DECIMAL(5,2),
  left_thigh_cm DECIMAL(5,2),
  right_thigh_cm DECIMAL(5,2),
  left_calf_cm DECIMAL(5,2),
  right_calf_cm DECIMAL(5,2),
  notes TEXT,
  log_date DATE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, log_date)
);

CREATE INDEX idx_body_measurement_logs_user_date ON body_measurement_logs(user_id, log_date);
