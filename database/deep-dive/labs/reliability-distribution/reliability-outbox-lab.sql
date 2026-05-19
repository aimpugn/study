-- DB deep-dive reliability/outbox lab.
-- PostgreSQL-oriented schema for DU34-DU37 and DU48.

DROP TABLE IF EXISTS lab_outbox;
DROP TABLE IF EXISTS lab_order;

CREATE TABLE lab_order (
    order_id bigserial PRIMARY KEY,
    customer_id integer NOT NULL,
    status text NOT NULL,
    shard_key integer NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE lab_outbox (
    outbox_id bigserial PRIMARY KEY,
    aggregate_type text NOT NULL,
    aggregate_id bigint NOT NULL,
    event_type text NOT NULL,
    payload text NOT NULL,
    claimed_at timestamp NULL,
    published_at timestamp NULL
);

BEGIN;
INSERT INTO lab_order(customer_id, status, shard_key)
VALUES (100, 'PAID', 100 % 16)
RETURNING order_id;

INSERT INTO lab_outbox(aggregate_type, aggregate_id, event_type, payload)
VALUES ('order', currval('lab_order_order_id_seq'), 'OrderPaid', '{"status":"PAID"}');
COMMIT;

-- DU48: worker claim pattern.
WITH claim AS (
    SELECT outbox_id
    FROM lab_outbox
    WHERE published_at IS NULL
      AND claimed_at IS NULL
    ORDER BY outbox_id
    LIMIT 10
    FOR UPDATE SKIP LOCKED
)
UPDATE lab_outbox o
SET claimed_at = CURRENT_TIMESTAMP
FROM claim
WHERE o.outbox_id = claim.outbox_id
RETURNING o.*;

-- DU34-DU35 observation anchors.
SELECT pg_current_wal_lsn();
-- On a replica, compare replay position with primary LSN using pg_last_wal_replay_lsn().

-- DU36-DU37 mental check.
SELECT shard_key, count(*)
FROM lab_order
GROUP BY shard_key
ORDER BY shard_key;
