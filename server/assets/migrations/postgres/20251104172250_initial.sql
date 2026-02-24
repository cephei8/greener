-- migrate:up

CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(128) NOT NULL,
    password_salt BYTEA NOT NULL,
    password_hash BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX ix_users_username ON users(username);


CREATE TABLE apikeys (
    id UUID PRIMARY KEY,
    description TEXT,
    secret_salt BYTEA NOT NULL,
    secret_hash BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX ix_apikeys_user_id ON apikeys(user_id);


CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    description TEXT,
    baggage JSON,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX ix_sessions_user_id ON sessions(user_id);


CREATE TABLE labels (
    id BIGSERIAL PRIMARY KEY,
    session_id UUID NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT,
    user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX ix_labels_session_id ON labels(session_id);
CREATE INDEX ix_labels_user_id ON labels(user_id);


CREATE TABLE testcases (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    classname TEXT,
    file TEXT,
    testsuite TEXT,
    output TEXT,
    status INTEGER NOT NULL,
    baggage JSON,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX ix_testcases_session_id ON testcases(session_id);
CREATE INDEX ix_testcases_user_id ON testcases(user_id);

-- migrate:down
