-- migrate:up

CREATE TABLE oauth_clients (
    id TEXT PRIMARY KEY,
    secret_hash BLOB,
    name TEXT NOT NULL,
    redirect_uris TEXT NOT NULL,
    user_id TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX ix_oauth_clients_user_id ON oauth_clients(user_id);

CREATE TABLE oauth_tokens (
    code TEXT PRIMARY KEY,
    access TEXT,
    refresh TEXT,
    client_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    redirect_uri TEXT,
    scope TEXT,
    code_challenge TEXT,
    code_challenge_method TEXT,
    code_expires_at TEXT,
    access_expires_at TEXT,
    refresh_expires_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (client_id) REFERENCES oauth_clients(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX ix_oauth_tokens_access ON oauth_tokens(access);
CREATE INDEX ix_oauth_tokens_refresh ON oauth_tokens(refresh);
CREATE INDEX ix_oauth_tokens_client_id ON oauth_tokens(client_id);
CREATE INDEX ix_oauth_tokens_user_id ON oauth_tokens(user_id);

-- migrate:down
