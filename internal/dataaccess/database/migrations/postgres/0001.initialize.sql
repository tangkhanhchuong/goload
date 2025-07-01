-- +migrate Up
CREATE TABLE IF NOT EXISTS accounts (
    id BIGSERIAL PRIMARY KEY,
    account_name VARCHAR(256) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS account_passwords (
    of_account_id BIGINT PRIMARY KEY,
    hash VARCHAR(128) NOT NULL,
    FOREIGN KEY (of_account_id) REFERENCES accounts(id)
);

CREATE TABLE IF NOT EXISTS public_keys (
    id BIGSERIAL PRIMARY KEY,
    public_key BYTEA NOT NULL
);

CREATE TABLE IF NOT EXISTS download_tasks (
    task_id BIGSERIAL PRIMARY KEY,
    of_account_id BIGINT,
    download_type SMALLINT NOT NULL,
    url TEXT NOT NULL,
    download_status SMALLINT NOT NULL,
    metadata TEXT NOT NULL,
    FOREIGN KEY (of_account_id) REFERENCES accounts(id)
);

-- +migrate Down
DROP TABLE IF EXISTS download_tasks;
DROP TABLE IF EXISTS token_public_keys;
DROP TABLE IF EXISTS account_passwords;
DROP TABLE IF EXISTS accounts;