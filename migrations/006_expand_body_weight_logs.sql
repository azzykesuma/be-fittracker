DO $$
DECLARE
  measurement_table TEXT;
BEGIN
  IF to_regclass('body_measurement_logs') IS NOT NULL THEN
    measurement_table := 'body_measurement_logs';
  ELSE
    measurement_table := 'body_weight_logs';
  END IF;

  EXECUTE format('ALTER TABLE %I
    ADD COLUMN IF NOT EXISTS bmi DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS body_fat_percentage DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS neck_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS shoulder_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS chest_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS waist_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS belly_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS hips_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS left_bicep_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS right_bicep_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS left_forearm_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS right_forearm_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS left_thigh_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS right_thigh_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS left_calf_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS right_calf_cm DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS notes TEXT', measurement_table);
END $$;
