DROP TABLE IF EXISTS habit_logs;
DROP TABLE IF EXISTS habits;

CREATE TABLE IF NOT EXISTS meal_logs (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  meal_date DATE NOT NULL,
  meal_type VARCHAR(20) NOT NULL,
  food_name VARCHAR(150) NOT NULL,
  calories INT NOT NULL,
  protein_g DECIMAL(6,2),
  carbs_g DECIMAL(6,2),
  fat_g DECIMAL(6,2),
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT meal_logs_meal_type_valid CHECK (meal_type IN ('breakfast', 'lunch', 'dinner', 'snack')),
  CONSTRAINT meal_logs_calories_non_negative CHECK (calories >= 0),
  CONSTRAINT meal_logs_protein_g_non_negative CHECK (protein_g IS NULL OR protein_g >= 0),
  CONSTRAINT meal_logs_carbs_g_non_negative CHECK (carbs_g IS NULL OR carbs_g >= 0),
  CONSTRAINT meal_logs_fat_g_non_negative CHECK (fat_g IS NULL OR fat_g >= 0)
);

CREATE INDEX IF NOT EXISTS idx_meal_logs_user_date ON meal_logs(user_id, meal_date);
CREATE INDEX IF NOT EXISTS idx_meal_logs_user_date_type ON meal_logs(user_id, meal_date, meal_type);
