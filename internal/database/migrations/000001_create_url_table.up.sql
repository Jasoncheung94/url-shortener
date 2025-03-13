CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,              -- Auto-incrementing primary key for each URL entry
    original_url TEXT NOT NULL,         -- The original URL (must be provided)
    short_url VARCHAR(255) NOT NULL,    -- The shortened URL (must be unique, assuming length constraint)
    custom_url VARCHAR(20),             -- The custom URL (optional, alphanumeric, length between 3-20)
    expiration_date TIMESTAMP,          -- The expiration date (optional)
    created_at TIMESTAMP DEFAULT NOW(), -- The creation date (automatically set to current time)
    CONSTRAINT unique_short_url UNIQUE (short_url) -- Ensure the short URL is unique
);
