CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS Photos (
    photo_id uuid PRIMARY KEY,
    owner_id integer references Users(user_id) NOT NULL,
    filepath VARCHAR(512) NOT NULL,
    thumbnail_filepath VARCHAR(512),
    thumbnail_job_id INTEGER,
    uploaded_timestamp timestamp NOT NULL,
    size integer NOT NULL,
    mime_type VARCHAR(64) NOT NULL,

    -- EXIF data, optional
    location GEOGRAPHY(Point, 4326),
    taken_timestamp timestamp,
    camera_model VARCHAR(128) 
)
