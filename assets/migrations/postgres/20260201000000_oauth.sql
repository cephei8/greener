-- migrate:up

CREATE TABLE oauth_clients (
    id VARCHAR(128) PRIMARY KEY,
    secret_hash BYTEA,
    name VARCHAR(255) NOT NULL,
    redirect_uris TEXT NOT NULL,
    user_id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX ix_oauth_clients_user_id ON oauth_clients(user_id);

CREATE TABLE oauth_tokens (
    code VARCHAR(512) PRIMARY KEY,
    access VARCHAR(512),
    refresh VARCHAR(512),
    client_id VARCHAR(128) NOT NULL,
    user_id UUID NOT NULL,
    redirect_uri TEXT,
    scope TEXT,
    code_challenge VARCHAR(128),
    code_challenge_method VARCHAR(16),
    code_expires_at TIMESTAMP WITH TIME ZONE,
    access_expires_at TIMESTAMP WITH TIME ZONE,
    refresh_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (client_id) REFERENCES oauth_clients(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX ix_oauth_tokens_access ON oauth_tokens(access);
CREATE INDEX ix_oauth_tokens_refresh ON oauth_tokens(refresh);
CREATE INDEX ix_oauth_tokens_client_id ON oauth_tokens(client_id);
CREATE INDEX ix_oauth_tokens_user_id ON oauth_tokens(user_id);

-- migrate:down
