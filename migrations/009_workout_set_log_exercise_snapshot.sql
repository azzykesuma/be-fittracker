ALTER TABLE workout_set_logs
  ADD COLUMN IF NOT EXISTS exercise_name VARCHAR(100);

UPDATE workout_set_logs AS logs
SET exercise_name = exercises.name
FROM exercises
WHERE logs.exercise_id = exercises.id
  AND logs.exercise_name IS NULL;

UPDATE workout_set_logs
SET exercise_name = 'Unknown exercise'
WHERE exercise_name IS NULL;

ALTER TABLE workout_set_logs
  ALTER COLUMN exercise_name SET NOT NULL,
  ALTER COLUMN exercise_id DROP NOT NULL;

ALTER TABLE workout_set_logs
  DROP CONSTRAINT IF EXISTS workout_set_logs_exercise_id_fkey;

ALTER TABLE workout_set_logs
  ADD CONSTRAINT workout_set_logs_exercise_id_fkey
  FOREIGN KEY (exercise_id) REFERENCES exercises(id) ON DELETE SET NULL;
