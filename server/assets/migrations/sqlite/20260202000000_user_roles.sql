-- migrate:up

ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'editor';

CREATE TABLE users_new (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    password_salt BLOB NOT NULL,
    password_hash BLOB NOT NULL,
    role TEXT NOT NULL DEFAULT 'viewer',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users_new SELECT id, username, password_salt, password_hash, role, created_at, updated_at FROM users;
DROP TABLE users;

ALTER TABLE users_new RENAME TO users;

CREATE UNIQUE INDEX ix_users_username ON users(username);

-- migrate:down
