CREATE TYPE job_status AS ENUM ('pending', 'in_progress', 'done', 'failed');

CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    status job_status NOT NULL DEFAULT 'pending', 
    payload JSONB,
    visible_at TIMESTAMP DEFAULT now(),
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
);
