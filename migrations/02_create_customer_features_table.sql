-- Feature customers: Although there are no queries that are driven
-- primarily by customer IDs, it is not difficult to imagine the
-- necessity for the application to view what features are active
-- for a customer.

-- I assume that for archived features, it is not necessary to maintain
-- the list of customer IDs it is active for.

CREATE TABLE customer_features
(
    id          BLOB PRIMARY KEY,
    customer_id TEXT NOT NULL,
    feature_id  BLOB NOT NULL,
    FOREIGN KEY (feature_id) REFERENCES features (id) ON DELETE CASCADE,
    UNIQUE (customer_id, feature_id)
);