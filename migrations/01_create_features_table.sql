CREATE TABLE features
(
    id             BLOB PRIMARY KEY,
    display_name   TEXT,
    technical_name TEXT      NOT NULL UNIQUE,
    expires_on     TIMESTAMP,
    description    TEXT,
    inverted       TINYINT   NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMP NOT NULL,
    updated_at     TIMESTAMP NOT NULL
);