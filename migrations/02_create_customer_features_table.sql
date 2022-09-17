-- Feature customers: Although there are no queries that are driven
-- primarily by customer IDs, it is not difficult to imagine the
-- necessity for the application to view what features are active
-- for a customer.

CREATE TABLE customer_features
(
    id          BLOB PRIMARY KEY,
    customer_id TEXT NOT NULL,
    feature_id  BLOB NOT NULL,
    FOREIGN     KEY (feature_id) REFERENCES feature(id) ON DELETE CASCADE
);