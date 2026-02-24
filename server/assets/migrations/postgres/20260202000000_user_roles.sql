-- migrate:up

ALTER TABLE users ADD COLUMN role VARCHAR(32) NOT NULL DEFAULT 'editor';
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'viewer';

-- migrate:down
