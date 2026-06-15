ALTER TABLE IF EXISTS body_weight_logs RENAME TO body_measurement_logs;

ALTER INDEX IF EXISTS idx_body_weight_logs_user_date RENAME TO idx_body_measurement_logs_user_date;
