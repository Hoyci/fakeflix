CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url TEXT NOT NULL,
    size_in_kb INTEGER NOT NULL,
    duration INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE thumbnails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE contents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    content_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE tv_shows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id UUID UNIQUE NOT NULL,
    thumbnail_id UUID UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_contents FOREIGN KEY(content_id) REFERENCES contents(id) ON DELETE CASCADE,
    CONSTRAINT fk_thumbnails FOREIGN KEY(thumbnail_id) REFERENCES thumbnails(id) ON DELETE SET NULL
);

CREATE TABLE episodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tv_show_id UUID NOT NULL,
    thumbnail_id UUID UNIQUE,
    video_id UUID UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    season INTEGER NOT NULL,
    number INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_tv_shows FOREIGN KEY(tv_show_id) REFERENCES tv_shows(id) ON DELETE CASCADE,
    CONSTRAINT fk_thumbnails FOREIGN KEY(thumbnail_id) REFERENCES thumbnails(id) ON DELETE SET NULL,
    CONSTRAINT fk_videos FOREIGN KEY(video_id) REFERENCES videos(id) ON DELETE CASCADE
);

CREATE TABLE movies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id UUID UNIQUE NOT NULL,
    video_id UUID UNIQUE NOT NULL,
    thumbnail_id UUID UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_contents FOREIGN KEY(content_id) REFERENCES contents(id) ON DELETE CASCADE,
    CONSTRAINT fk_videos FOREIGN KEY(video_id) REFERENCES videos(id) ON DELETE CASCADE,
    CONSTRAINT fk_thumbnails FOREIGN KEY(thumbnail_id) REFERENCES thumbnails(id) ON DELETE SET NULL
);

CREATE INDEX idx_contents_deleted_at ON contents(deleted_at);
CREATE INDEX idx_movies_deleted_at ON movies(deleted_at);
CREATE INDEX idx_tv_shows_deleted_at ON tv_shows(deleted_at);
CREATE INDEX idx_episodes_deleted_at ON episodes(deleted_at);
CREATE INDEX idx_videos_deleted_at ON videos(deleted_at);
CREATE INDEX idx_thumbnails_deleted_at ON thumbnails(deleted_at);
