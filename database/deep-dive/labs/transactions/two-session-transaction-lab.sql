-- DB deep-dive two-session transaction lab.
-- Copy Session A and Session B blocks into two separate SQL consoles.

DROP TABLE IF EXISTS lab_account;
CREATE TABLE lab_account (
    account_id integer PRIMARY KEY,
    balance integer NOT NULL,
    on_call boolean NOT NULL DEFAULT true
);

INSERT INTO lab_account(account_id, balance, on_call) VALUES
    (1, 100, true),
    (2, 100, true);

-- DU25-DU26: transaction boundary and savepoint.
BEGIN;
UPDATE lab_account SET balance = balance - 10 WHERE account_id = 1;
SAVEPOINT after_debit;
UPDATE lab_account SET balance = balance + 10 WHERE account_id = 2;
-- ROLLBACK TO SAVEPOINT after_debit;
COMMIT;

-- DU30-DU31: nonrepeatable read / snapshot experiment.
-- Session A
-- BEGIN;
-- SET TRANSACTION ISOLATION LEVEL READ COMMITTED;
-- SELECT balance FROM lab_account WHERE account_id = 1;
-- SELECT balance FROM lab_account WHERE account_id = 1;
-- COMMIT;

-- Session B, between A's two SELECTs
-- BEGIN;
-- UPDATE lab_account SET balance = balance + 1 WHERE account_id = 1;
-- COMMIT;

-- DU32-DU33: deadlock experiment.
-- Session A
-- BEGIN;
-- UPDATE lab_account SET balance = balance - 1 WHERE account_id = 1;
-- UPDATE lab_account SET balance = balance + 1 WHERE account_id = 2;
-- COMMIT;

-- Session B, opposite order
-- BEGIN;
-- UPDATE lab_account SET balance = balance - 1 WHERE account_id = 2;
-- UPDATE lab_account SET balance = balance + 1 WHERE account_id = 1;
-- COMMIT;

-- PostgreSQL observation helpers.
SELECT pid, state, wait_event_type, wait_event, query
FROM pg_stat_activity
WHERE datname = current_database();

SELECT locktype, relation::regclass, mode, granted
FROM pg_locks
WHERE pid = pg_backend_pid();
