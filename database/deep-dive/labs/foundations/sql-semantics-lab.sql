-- DB deep-dive foundations lab.
-- Run in a disposable PostgreSQL-compatible schema unless noted otherwise.

DROP TABLE IF EXISTS lab_order_line;
DROP TABLE IF EXISTS lab_customer;

CREATE TABLE lab_customer (
    customer_id integer PRIMARY KEY,
    email text UNIQUE,
    display_name text NOT NULL
);

CREATE TABLE lab_order_line (
    order_id integer NOT NULL,
    line_no integer NOT NULL,
    customer_id integer NOT NULL REFERENCES lab_customer(customer_id),
    sku text NOT NULL,
    qty integer NOT NULL,
    shipped_at timestamp NULL,
    PRIMARY KEY (order_id, line_no)
);

INSERT INTO lab_customer(customer_id, email, display_name) VALUES
    (1, 'kim@example.test', 'kim'),
    (2, NULL, 'anonymous-a'),
    (3, NULL, 'anonymous-b');

INSERT INTO lab_order_line(order_id, line_no, customer_id, sku, qty, shipped_at) VALUES
    (10, 1, 1, 'BOOK', 2, NULL),
    (10, 2, 1, 'PEN', 1, CURRENT_TIMESTAMP),
    (11, 1, 2, 'BOOK', 1, NULL);

-- DU03-DU04: relation/key/join and logical SELECT order.
SELECT c.customer_id, c.display_name, sum(l.qty) AS total_qty
FROM lab_customer c
JOIN lab_order_line l ON l.customer_id = c.customer_id
WHERE l.qty > 0
GROUP BY c.customer_id, c.display_name
HAVING sum(l.qty) >= 1
ORDER BY total_qty DESC, c.customer_id;

-- DU05: NULL and UNKNOWN. The NOT IN query should surprise you when NULL participates.
SELECT customer_id, email, email = NULL AS equals_null
FROM lab_customer;

SELECT customer_id
FROM lab_customer
WHERE email NOT IN ('kim@example.test', NULL);

SELECT customer_id
FROM lab_customer
WHERE email IS DISTINCT FROM 'kim@example.test';

-- DU07: prepared statement boundary. Compare parameter binding with string concatenation in your client code.
PREPARE find_customer(text) AS
SELECT customer_id, display_name
FROM lab_customer
WHERE email = $1;

EXECUTE find_customer('kim@example.test');
DEALLOCATE find_customer;
