CREATE TABLE IF NOT EXISTS messaging_user_public_keys (
    user_id UUID PRIMARY KEY,
    public_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS messaging_thread_keys (
    thread_id UUID NOT NULL REFERENCES messaging_threads(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    encrypted_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (thread_id, user_id)
);
