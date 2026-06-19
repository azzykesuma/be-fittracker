ALTER TABLE users ADD COLUMN gender VARCHAR(10) NOT NULL DEFAULT 'male';
ALTER TABLE users ADD CONSTRAINT users_gender_check CHECK (gender IN ('male', 'female'));
