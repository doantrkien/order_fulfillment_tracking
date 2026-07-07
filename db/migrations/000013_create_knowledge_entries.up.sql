CREATE TABLE knowledge_entries (
    id         SERIAL PRIMARY KEY,
    slug       VARCHAR(64) UNIQUE NOT NULL,
    title      VARCHAR(256) NOT NULL,
    body       TEXT NOT NULL,
    is_active  BOOLEAN DEFAULT true NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
