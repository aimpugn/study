CREATE TABLE IF NOT EXISTS item_keys (
    id SERIAL PRIMARY KEY,
    item_key VARCHAR(255) NOT NULL UNIQUE,
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    modified TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);
CREATE INDEX idx_item_keys_key ON item_keys(item_key);

INSERT INTO item_keys (item_key, created, modified) VALUES ('some_key', NOW(), NOW());

CREATE TABLE IF NOT EXISTS item_values (
    id SERIAL PRIMARY KEY,
    item_key_id INT NOT NULL,
    item_value VARCHAR(255) NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    modified TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (item_key_id) REFERENCES item_keys(id)
);
CREATE INDEX idx_item_values_item_value ON item_values(item_value);

INSERT INTO item_values (item_key_id, item_value, created, modified) VALUES (1, 'some_value1', NOW(), NOW());
INSERT INTO item_values (item_key_id, item_value, created, modified) VALUES (1, 'some_value2', NOW(), NOW());

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    modified TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS user_accounts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    account_id INT NOT NULL,
    roles JSON DEFAULT '[]' NOT NULL,
    active BOOLEAN NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    modified TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    balance DECIMAL DEFAULT 0 NOT NULL,
    currency VARCHAR(10) NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    modified TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS account_transactions (
    id SERIAL PRIMARY KEY,
    from_account_id INT NOT NULL,
    to_account_id INT NOT NULL,
    amount DECIMAL DEFAULT 0 NOT NULL,
    currency VARCHAR(10) NOT NULL,
    description VARCHAR(1024) NOT NULL,
    transaction_type VARCHAR(10) NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (from_account_id) REFERENCES accounts(id),
    FOREIGN KEY (to_account_id) REFERENCES accounts(id)
);