CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL   PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    name            TEXT        NOT NULL,
    email           VARCHAR(255),
    email_verified  BOOLEAN     NOT NULL DEFAULT FALSE,
    mobile          VARCHAR(255),
    mobile_verified BOOLEAN     NOT NULL DEFAULT FALSE,
    password        TEXT        NOT NULL,
    role            INTEGER     NOT NULL DEFAULT 1,
    google_id       TEXT,
    google_avatar   TEXT
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique  ON users (email)  WHERE email  IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_mobile_unique ON users (mobile) WHERE mobile IS NOT NULL;

CREATE TABLE IF NOT EXISTS verify (
    id         BIGSERIAL   PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    target     TEXT        NOT NULL,
    token      TEXT        NOT NULL,
    verified   BOOLEAN     NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT verify_target_unique UNIQUE (target),
    CONSTRAINT verify_token_unique  UNIQUE (token)
);

CREATE INDEX IF NOT EXISTS idx_verify_deleted_at ON verify (deleted_at);
