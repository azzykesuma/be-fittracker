CREATE TABLE refresh_sessions (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  rotated_at TIMESTAMPTZ
);

CREATE INDEX idx_refresh_sessions_user_id ON refresh_sessions(user_id);
CREATE INDEX idx_refresh_sessions_expires_at ON refresh_sessions(expires_at);
