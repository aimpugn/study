-- Application boundary, idempotency, money, and financial-state lab.
-- PostgreSQL-oriented, but the modeling ideas transfer to other RDBMSs.

DROP TABLE IF EXISTS lab_payment_event;
DROP TABLE IF EXISTS lab_payment_request;
DROP TABLE IF EXISTS lab_ledger_entry;

CREATE TABLE lab_payment_request (
    idempotency_key text PRIMARY KEY,
    request_hash text NOT NULL,
    status text NOT NULL,
    response_body text NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE lab_ledger_entry (
    ledger_id bigserial PRIMARY KEY,
    account_id bigint NOT NULL,
    direction text NOT NULL CHECK (direction IN ('DEBIT', 'CREDIT')),
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    currency char(3) NOT NULL,
    external_ref text NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE lab_payment_event (
    event_id bigserial PRIMARY KEY,
    payment_id text NOT NULL,
    event_type text NOT NULL,
    raw_status text NOT NULL,
    observed_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- DU44: winner election by idempotency key.
INSERT INTO lab_payment_request(idempotency_key, request_hash, status)
VALUES ('pay-2026-0001', 'hash-a', 'PROCESSING')
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING *;

SELECT *
FROM lab_payment_request
WHERE idempotency_key = 'pay-2026-0001';

-- DU45: keep money in minor units at the DB boundary.
INSERT INTO lab_ledger_entry(account_id, direction, amount_minor, currency, external_ref)
VALUES
    (1, 'DEBIT', 1000, 'KRW', 'pay-2026-0001'),
    (2, 'CREDIT', 1000, 'KRW', 'pay-2026-0001');

SELECT external_ref,
       sum(CASE WHEN direction = 'DEBIT' THEN amount_minor ELSE -amount_minor END) AS debit_minus_credit
FROM lab_ledger_entry
GROUP BY external_ref;

-- DU46-DU47: unknown external state is not the same as failure.
INSERT INTO lab_payment_event(payment_id, event_type, raw_status)
VALUES
    ('pay-2026-0001', 'API_TIMEOUT', 'UNKNOWN'),
    ('pay-2026-0001', 'WEBHOOK', 'SUCCEEDED'),
    ('pay-2026-0001', 'SETTLEMENT_FILE', 'SETTLED');

SELECT payment_id, array_agg(event_type || ':' || raw_status ORDER BY observed_at, event_id) AS recovery_trace
FROM lab_payment_event
GROUP BY payment_id;
