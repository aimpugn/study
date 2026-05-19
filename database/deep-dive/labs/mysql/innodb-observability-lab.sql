-- MySQL/InnoDB observability lab for DU38-DU39.
-- Run only in a disposable MySQL schema.

CREATE TABLE IF NOT EXISTS lab_innodb_order (
    order_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    PRIMARY KEY (order_id),
    KEY idx_status (status),
    KEY idx_customer_status (customer_id, status)
) ENGINE=InnoDB;

INSERT INTO lab_innodb_order(order_id, customer_id, status, amount)
VALUES
    (1, 10, 'READY', 100.00),
    (2, 10, 'PAID', 200.00)
ON DUPLICATE KEY UPDATE status = VALUES(status);

EXPLAIN SELECT order_id
FROM lab_innodb_order
WHERE customer_id = 10 AND status = 'READY';

SHOW ENGINE INNODB STATUS;

SELECT *
FROM information_schema.innodb_trx;

SELECT *
FROM performance_schema.data_locks;

-- Online DDL / metadata lock observation should be done in separate sessions.
-- Session A: START TRANSACTION; SELECT * FROM lab_innodb_order WHERE order_id = 1;
-- Session B: ALTER TABLE lab_innodb_order ADD COLUMN lab_note VARCHAR(30) NULL, ALGORITHM=INPLACE, LOCK=NONE;
-- Session C: SHOW PROCESSLIST;
