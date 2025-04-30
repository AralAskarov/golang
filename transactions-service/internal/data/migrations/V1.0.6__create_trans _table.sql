CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    uuid BYTEA NOT NULL,
    amount INT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('deposit', 'withdrawal')),
    time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    uuid TEXT,
    balance INT
);