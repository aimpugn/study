# Procedure

## examples

```sql
CREATE PROCEDURE insert_demo_that_table_by_rody(IN some_table_id INT, IN cancels_cnt INT)
BEGIN
    DECLARE some_table_cnt INT;
    DECLARE i INT DEFAULT 1; -- notice 1개 감안
    SELECT COUNT(1) INTO some_table_cnt FROM some_table p WHERE p.id = some_table_id;
    SELECT some_table_cnt, cancels_cnt;
    IF some_table_cnt > 0 THEN
        START TRANSACTION;

        INSERT INTO that_table (col1, col2, ... colN) VALUES
            (..., ..., ...);

        WHILE i < cancels_cnt DO
            INSERT INTO that_table (col1, col2, col3, ...) VALUES
                (..., ..., ...);
            SET i = i + 1;
        END WHILE;
        COMMIT;
    END IF;

    SELECT COUNT(1) AS notices_cnt FROM that_table pn WHERE pn.some_table_id = some_table_id;
END;

CREATE PROCEDURE insert_demo_some_table_by_rody(IN dup_uid_from_client VARCHAR(40) CHARACTER SET utf8 COLLATE utf8_unicode_ci, IN loop_cnt INT)
BEGIN
    DECLARE some_table_cnt INT;
    DECLARE status enum('enum1','enum2','enum3','enum4','enum5');
    DECLARE i INT DEFAULT 0;
    SELECT COUNT(1) INTO some_table_cnt FROM some_table p WHERE p.uid_from_client = dup_uid_from_client COLLATE utf8_unicode_ci;

    SELECT some_table_cnt, loop_cnt;
    IF loop_cnt > 0 THEN
        START TRANSACTION;
        WHILE i < loop_cnt AND i < 100 DO
            SET @imp_uid = CONCAT('i_will_delete_it-', ROUND(RAND() * 100000));
            SET @created_at = DATE_ADD('2023-09-07 16:58:47', INTERVAL FLOOR(48 * RAND()) HOUR);
            SET @status = IF(RAND() * 10 < 8, 'enum2', 'enum3');
            SET @cancelled_at = IF(@status = 'enum2', NULL, DATE_ADD(@created_at, INTERVAL FLOOR(48 * RAND()) MINUTE));
            SET @paid_at = DATE_ADD(@created_at, INTERVAL FLOOR(48 * RAND()) MINUTE);
            SET @cancel_reason = IF(@status = 'enum2', NULL, 'cancel');


            INSERT INTO some_table (col1, col2, ..., colN)
            VALUES (@status,@paid_at, NULL, @cancelled_at, NULL, @cancel_reason, @paid_at, @cancelled_at, NULL, @created_at, @created_at);

            SET i = i + 1;
        END WHILE;
        COMMIT;
    END IF;

    SELECT COUNT(1) AS result_cnt FROM some_table p WHERE p.uid_from_client = dup_uid_from_client;
    SELECT p.id FROM some_table p WHERE p.uid_from_client = dup_uid_from_client;
END;

DROP PROCEDURE IF EXISTS insert_demo_this_table_by_rody;

CREATE PROCEDURE insert_demo_this_table_by_rody(IN some_table_id INT, IN loop_cnt INT, IN cancel_amount INT)
BEGIN
    DECLARE some_table_cnt INT;
    DECLARE i INT DEFAULT 0;
    SELECT COUNT(1) INTO some_table_cnt FROM some_table p WHERE p.id = some_table_id;
    SELECT some_table_cnt, loop_cnt;
    IF some_table_cnt > 0 AND loop_cnt > 0 THEN
        START TRANSACTION;
        WHILE i < loop_cnt DO
            INSERT INTO this_table (col1, col2, ... colN) VALUES
                (..., ..., ...);
            SET i = i + 1;
        END WHILE;
        COMMIT;
    END IF;

    SELECT COUNT(1) AS cancels_cnt FROM this_table pc WHERE pc.some_table_id = some_table_id;
END;

DROP PROCEDURE IF EXISTS delete_demo_data_by_rody;

CREATE PROCEDURE delete_demo_data_by_rody()
BEGIN
    DECLARE done INT DEFAULT 0;
    DECLARE current_some_table_id INT;
    DECLARE some_table_id_cursor CURSOR FOR SELECT p.id FROM some_table p WHERE p.uid_from_client = 'rody_large_transactions_test';
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1;

    START TRANSACTION;
    OPEN some_table_id_cursor;

    read_loop: LOOP
        FETCH some_table_id_cursor INTO current_some_table_id;
        IF done THEN
          LEAVE read_loop;
        END IF;

        IF current_some_table_id >= 370075616 THEN
            DELETE FROM some_table WHERE id = current_some_table_id;
            DELETE FROM this_table WHERE some_table_id = current_some_table_id;
            DELETE FROM that_table WHERE some_table_id = current_some_table_id;
        END IF;
    END LOOP;

    CLOSE some_table_id_cursor;

    COMMIT;
END;


SHOW PROCEDURE STATUS;

DROP PROCEDURE IF EXISTS insert_demo_some_table_by_rody;
DROP PROCEDURE IF EXISTS delete_demo_data_by_rody;
DROP PROCEDURE IF EXISTS insert_demo_this_table_by_rody;
DROP PROCEDURE IF EXISTS insert_demo_that_table_by_rody;
```
