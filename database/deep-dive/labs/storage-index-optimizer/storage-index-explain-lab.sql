-- DB deep-dive storage/index/optimizer lab.
-- PostgreSQL-oriented. For MySQL, translate EXPLAIN ANALYZE to EXPLAIN and SHOW statements.

DROP TABLE IF EXISTS lab_event;

CREATE TABLE lab_event (
    event_id bigserial PRIMARY KEY,
    tenant_id integer NOT NULL,
    event_type text NOT NULL,
    occurred_at timestamp NOT NULL,
    payload text NOT NULL
);

INSERT INTO lab_event(tenant_id, event_type, occurred_at, payload)
SELECT
    (g % 10) + 1,
    CASE WHEN g % 20 = 0 THEN 'PAYMENT_FAILED' ELSE 'PAYMENT_OK' END,
    CURRENT_TIMESTAMP - (g || ' minutes')::interval,
    repeat('x', 100)
FROM generate_series(1, 5000) AS g;

ANALYZE lab_event;

-- DU17-DU22: plan before index.
EXPLAIN (ANALYZE, BUFFERS)
SELECT event_id, occurred_at
FROM lab_event
WHERE tenant_id = 3
  AND event_type = 'PAYMENT_FAILED'
ORDER BY occurred_at DESC
LIMIT 20;

CREATE INDEX lab_event_tenant_type_time_idx
ON lab_event(tenant_id, event_type, occurred_at DESC);

ANALYZE lab_event;

-- DU14-DU16, DU20-DU22: plan after composite index.
EXPLAIN (ANALYZE, BUFFERS)
SELECT event_id, occurred_at
FROM lab_event
WHERE tenant_id = 3
  AND event_type = 'PAYMENT_FAILED'
ORDER BY occurred_at DESC
LIMIT 20;

-- DU19: OFFSET cost shape. Increase the OFFSET and observe row work.
EXPLAIN (ANALYZE, BUFFERS)
SELECT event_id, occurred_at
FROM lab_event
WHERE tenant_id = 3
ORDER BY occurred_at DESC
OFFSET 200 LIMIT 20;
