-- === MIGRATION SQL ===
-- 1. novel_schedules table
CREATE TABLE IF NOT EXISTS novel_schedules (
    id SERIAL PRIMARY KEY,
    novel_id INTEGER REFERENCES novels(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT TRUE,
    interval_seconds INT NOT NULL DEFAULT 86400,
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- 2. chapter_crawl_logs table (optional, for auditing)
CREATE TABLE IF NOT EXISTS chapter_crawl_logs (
    id SERIAL PRIMARY KEY,
    chapter_id INTEGER REFERENCES chapters(id) ON DELETE CASCADE,
    status TEXT CHECK (status IN ('success', 'failed')) NOT NULL,
    error TEXT,
    created_at TIMESTAMP DEFAULT now()
);

-- 3. Add column to novels for metadata refresh indicator
ALTER TABLE novels
ADD COLUMN IF NOT EXISTS need_update BOOLEAN DEFAULT FALSE;

-- 4. Remove unused columns
ALTER TABLE public.novels DROP COLUMN author;
ALTER TABLE public.novels DROP COLUMN genres;
ALTER TABLE public.novels DROP COLUMN cover_url;
ALTER TABLE public.novels DROP COLUMN description;
