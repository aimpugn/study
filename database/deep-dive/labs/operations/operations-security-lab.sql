-- Operations and security lab for DU49-DU52.
-- PostgreSQL-oriented. Adapt privilege syntax for MySQL when needed.

DROP TABLE IF EXISTS lab_sensitive_order;

CREATE TABLE lab_sensitive_order (
    order_id bigserial PRIMARY KEY,
    tenant_id integer NOT NULL,
    owner_name text NOT NULL,
    amount numeric(12,2) NOT NULL
);

INSERT INTO lab_sensitive_order(tenant_id, owner_name, amount)
VALUES (1, 'kim', 100.00), (2, 'lee', 200.00);

-- DU49-DU50: symptom-oriented observation.
EXPLAIN (ANALYZE, BUFFERS)
SELECT *
FROM lab_sensitive_order
WHERE tenant_id = 1;

SELECT pid, state, wait_event_type, wait_event, query
FROM pg_stat_activity
WHERE datname = current_database();

SELECT relname, seq_scan, idx_scan, n_live_tup, n_dead_tup
FROM pg_stat_all_tables
WHERE relname = 'lab_sensitive_order';

-- DU51-DU52: least privilege and row-level security sketch.
-- Run role statements only in a disposable database with sufficient privileges.
-- CREATE ROLE lab_app LOGIN PASSWORD 'change-me';
-- GRANT SELECT ON lab_sensitive_order TO lab_app;
-- ALTER TABLE lab_sensitive_order ENABLE ROW LEVEL SECURITY;
-- CREATE POLICY tenant_1_policy ON lab_sensitive_order
--     USING (tenant_id = current_setting('app.tenant_id')::integer);
-- SET app.tenant_id = '1';
-- SELECT * FROM lab_sensitive_order;
