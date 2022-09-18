CREATE TABLE archived_features
(
    id             BLOB PRIMARY KEY,
    display_name   TEXT,
    technical_name TEXT NOT NULL,
    description    TEXT,
    created_at     TIMESTAMP NOT NULL,
    updated_at     TIMESTAMP NOT NULL
);
