
CREATE TABLE generated_clips (
    id BIGSERIAL PRIMARY KEY,

    video_id BIGINT NOT NULL UNIQUE
        REFERENCES videos(id) ON DELETE CASCADE,

    start_time BIGINT NOT NULL,
    end_time BIGINT NOT NULL,

    ai_score INT NOT NULL,

    title VARCHAR(100) NOT NULL,
    caption TEXT NOT NULL,

    file_path TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_clip_duration
        CHECK (start_time >= 0 AND end_time > start_time),

    CONSTRAINT valid_ai_score
        CHECK (ai_score BETWEEN 0 AND 40)
);
