CREATE TABLE IF NOT EXISTS migrate_test_users (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);
