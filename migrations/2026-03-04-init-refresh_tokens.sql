CREATE TABLE refresh_tokens (
    token TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCE users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL
);