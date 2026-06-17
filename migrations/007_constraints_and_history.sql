ALTER TABLE IF EXISTS body_weight_logs RENAME TO body_measurement_logs;

ALTER INDEX IF EXISTS idx_body_weight_logs_user_date RENAME TO idx_body_measurement_logs_user_date;

ALTER TABLE users
  ADD CONSTRAINT users_height_cm_positive CHECK (height_cm IS NULL OR height_cm > 0),
  ADD CONSTRAINT users_weight_kg_positive CHECK (weight_kg IS NULL OR weight_kg > 0);

ALTER TABLE exercises
  ADD CONSTRAINT exercises_target_sets_positive CHECK (target_sets IS NULL OR target_sets > 0),
  ADD CONSTRAINT exercises_target_reps_positive CHECK (target_reps IS NULL OR target_reps > 0),
  ADD CONSTRAINT exercises_target_weight_kg_non_negative CHECK (target_weight_kg IS NULL OR target_weight_kg >= 0),
  ADD CONSTRAINT exercises_rest_seconds_non_negative CHECK (rest_seconds >= 0);

ALTER TABLE workout_sessions
  ADD CONSTRAINT workout_sessions_status_valid CHECK (status IN ('in_progress', 'finished', 'cancelled'));

ALTER TABLE workout_set_logs
  ADD CONSTRAINT workout_set_logs_set_number_positive CHECK (set_number > 0),
  ADD CONSTRAINT workout_set_logs_reps_non_negative CHECK (reps >= 0),
  ADD CONSTRAINT workout_set_logs_weight_kg_non_negative CHECK (weight_kg IS NULL OR weight_kg >= 0);

ALTER TABLE body_measurement_logs
  ADD CONSTRAINT body_measurement_logs_weight_kg_positive CHECK (weight_kg > 0),
  ADD CONSTRAINT body_measurement_logs_bmi_positive CHECK (bmi IS NULL OR bmi > 0),
  ADD CONSTRAINT body_measurement_logs_body_fat_percentage_valid CHECK (body_fat_percentage IS NULL OR body_fat_percentage BETWEEN 0 AND 100),
  ADD CONSTRAINT body_measurement_logs_neck_cm_positive CHECK (neck_cm IS NULL OR neck_cm > 0),
  ADD CONSTRAINT body_measurement_logs_shoulder_cm_positive CHECK (shoulder_cm IS NULL OR shoulder_cm > 0),
  ADD CONSTRAINT body_measurement_logs_chest_cm_positive CHECK (chest_cm IS NULL OR chest_cm > 0),
  ADD CONSTRAINT body_measurement_logs_waist_cm_positive CHECK (waist_cm IS NULL OR waist_cm > 0),
  ADD CONSTRAINT body_measurement_logs_belly_cm_positive CHECK (belly_cm IS NULL OR belly_cm > 0),
  ADD CONSTRAINT body_measurement_logs_hips_cm_positive CHECK (hips_cm IS NULL OR hips_cm > 0),
  ADD CONSTRAINT body_measurement_logs_left_bicep_cm_positive CHECK (left_bicep_cm IS NULL OR left_bicep_cm > 0),
  ADD CONSTRAINT body_measurement_logs_right_bicep_cm_positive CHECK (right_bicep_cm IS NULL OR right_bicep_cm > 0),
  ADD CONSTRAINT body_measurement_logs_left_forearm_cm_positive CHECK (left_forearm_cm IS NULL OR left_forearm_cm > 0),
  ADD CONSTRAINT body_measurement_logs_right_forearm_cm_positive CHECK (right_forearm_cm IS NULL OR right_forearm_cm > 0),
  ADD CONSTRAINT body_measurement_logs_left_thigh_cm_positive CHECK (left_thigh_cm IS NULL OR left_thigh_cm > 0),
  ADD CONSTRAINT body_measurement_logs_right_thigh_cm_positive CHECK (right_thigh_cm IS NULL OR right_thigh_cm > 0),
  ADD CONSTRAINT body_measurement_logs_left_calf_cm_positive CHECK (left_calf_cm IS NULL OR left_calf_cm > 0),
  ADD CONSTRAINT body_measurement_logs_right_calf_cm_positive CHECK (right_calf_cm IS NULL OR right_calf_cm > 0);
