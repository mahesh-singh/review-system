CREATE TABLE IF NOT EXISTS processed_files (
    id SERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    s3path TEXT NOT NULL,
    processed_at TIMESTAMPTZ,
    records_count INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);