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
    item_key_id BIGINT NOT NULL,
    item_value VARCHAR(255) NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    modified TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (item_key_id) REFERENCES item_keys(id)
);
CREATE INDEX idx_item_values_item_value ON item_values(item_value);

INSERT INTO item_values (item_key_id, item_value, created, modified) VALUES (1, 'some_value1', NOW(), NOW());
INSERT INTO item_values (item_key_id, item_value, created, modified) VALUES (1, 'some_value2', NOW(), NOW());
