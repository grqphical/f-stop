CREATE TABLE IF NOT EXISTS Tags (
    id SERIAL PRIMARY KEY,
    owner_id INTEGER REFERENCES Users(user_id) NOT NULL,
    name TEXT NOT NULL,
    UNIQUE(owner_id, name)
)

CREATE TABLE IF NOT EXISTS TagAssignments (
    photo_id UUID REFERENCES Photos(photo_id) ON DELETE CASCADE NOT NULL,
    tag_id UUID REFERENCES Tags(id) ON DELETE CASCADE NOT NULL,
    PRIMARY KEY (photo_id, tag_id)
)
