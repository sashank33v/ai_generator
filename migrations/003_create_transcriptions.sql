CREATE TABLE transcriptions (
    id BIGSERIAL PRIMARY KEY,
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    language VARCHAR(20) NOT NULL,
    start_time BIGINT NOT NULL,
    end_time BIGINT NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transcriptions_video_id
ON transcriptions(video_id);