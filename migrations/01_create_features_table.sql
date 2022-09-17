-- Feature archival: I've opted for an `archived` field for simplicity.
-- If the features table were to grow large enough to cause issues
-- querying it, one could create a separate `archived_features` table,
-- and implement the archival of a feature as a transaction that
-- removes the feature from `features`, and inserts it into
-- `archived_features`.

CREATE TABLE features
(
    id             BLOB PRIMARY KEY,
    display_name   TEXT,
    technical_name TEXT NOT NULL,
    expires_on     TIMESTAMP,
    description    TEXT,
    inverted       TINYINT,
    archived       TINYINT
);