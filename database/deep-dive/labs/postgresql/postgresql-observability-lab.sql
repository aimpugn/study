-- PostgreSQL observability lab for DU40-DU41.

DROP TABLE IF EXISTS lab_pg_order;

CREATE TABLE lab_pg_order (
    order_id bigserial PRIMARY KEY,
    customer_id bigint NOT NULL,
    status text NOT NULL,
    amount numeric(12,2) NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX lab_pg_order_customer_status_idx
ON lab_pg_order(customer_id, status);

INSERT INTO lab_pg_order(customer_id, status, amount)
SELECT g % 20, CASE WHEN g % 3 = 0 THEN 'PAID' ELSE 'READY' END, (g % 100)::numeric
FROM generate_series(1, 2000) AS g;

ANALYZE lab_pg_order;

EXPLAIN (ANALYZE, BUFFERS)
SELECT order_id
FROM lab_pg_order
WHERE customer_id = 7 AND status = 'PAID';

SELECT ctid, xmin, xmax, order_id, status
FROM lab_pg_order
WHERE order_id <= 5;

SELECT relname, n_live_tup, n_dead_tup, vacuum_count, autovacuum_count
FROM pg_stat_all_tables
WHERE relname = 'lab_pg_order';

SELECT locktype, mode, granted, relation::regclass
FROM pg_locks
WHERE relation = 'lab_pg_order'::regclass;

SELECT pg_current_wal_lsn();
