-- 1. Websites table
CREATE TABLE websites (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    base_url TEXT NOT NULL,
    script_name TEXT NOT NULL,
    crawl_interval INT DEFAULT 86400,
    username TEXT,
    password TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT now()
);

-- 2. Novels table
CREATE TABLE novels (
    id SERIAL PRIMARY KEY,
    website_id INTEGER REFERENCES websites(id),
    external_id TEXT,
    title TEXT NOT NULL,
    author TEXT,
    genres TEXT[],
    cover_url TEXT,
    description TEXT,
    status TEXT CHECK (status IN ('ongoing', 'completed')) DEFAULT 'ongoing',
    source_url TEXT NOT NULL,
    last_crawled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT now()
);

-- 3. Chapters table
CREATE TABLE chapters (
    id SERIAL PRIMARY KEY,
    novel_id INTEGER REFERENCES novels(id) ON DELETE CASCADE,
    external_id TEXT,
    title TEXT,
    chapter_number INT,
    url TEXT NOT NULL,
    content TEXT,
    status TEXT CHECK (status IN ('pending', 'crawled', 'failed')) DEFAULT 'pending',
    crawled_at TIMESTAMP,
    error TEXT
);

-- 4. Crawl Jobs
CREATE TABLE crawl_jobs (
    id SERIAL PRIMARY KEY,
    novel_id INTEGER REFERENCES novels(id),
    status TEXT CHECK (status IN ('pending', 'in_progress', 'success', 'failed')) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT now(),
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    error TEXT
);

-- 5. Agents
CREATE TABLE agents (
    id UUID PRIMARY KEY,
    name TEXT,
    ip_address TEXT,
    last_heartbeat TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT now()
);

-- 6. Users (for API access control)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT now()
);

-- 7. API Tokens
CREATE TABLE api_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    description TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT now(),
    last_used_at TIMESTAMP
);
