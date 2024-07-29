# Procedure

## examples

```sql
CREATE PROCEDURE insert_demo_payment_notices_by_rody(IN payment_id INT, IN cancels_cnt INT)
BEGIN
    DECLARE payments_cnt INT;
    DECLARE i INT DEFAULT 1; -- notice 1개 감안
    SELECT COUNT(1) INTO payments_cnt FROM payments p WHERE p.id = payment_id;
    SELECT payments_cnt, cancels_cnt;
    IF payments_cnt > 0 THEN
        START TRANSACTION;
        INSERT INTO payment_notices (payment_id,request_id,status_code,notice_url,notice_type,to_merchant,from_merchant,header_from,created) VALUES
            (payment_id,ROUND(RAND() * 100000000),200,'https://www.service.domain.kr','notice','imp_uid=imp_rody_will_delete_it&merchant_uid=rody_large_transactions_test&status=paid','{"success":"true"}','[0] ', NOW());
    
        WHILE i < cancels_cnt DO
            INSERT INTO payment_notices (payment_id,request_id,status_code,notice_url,notice_type,to_merchant,from_merchant,header_from,created) VALUES
                (payment_id,ROUND(RAND() * 100000000),200,'https://www.service.domain.kr','cancel','imp_uid=imp_rody_will_delete_it&merchant_uid=rody_large_transactions_test&status=cancelled','{"success":"true"}','[0] ', NOW());
            SET i = i + 1;
        END WHILE;
        COMMIT;
    END IF;

    SELECT COUNT(1) AS notices_cnt FROM payment_notices pn WHERE pn.payment_id = payment_id;
END;

CREATE PROCEDURE insert_demo_payments_by_rody(IN dup_merchant_uid VARCHAR(40) CHARACTER SET utf8 COLLATE utf8_unicode_ci, IN loop_cnt INT)
BEGIN
    DECLARE payments_cnt INT;
    DECLARE status enum('ready','paid','cancelled','failed','pending');
    DECLARE i INT DEFAULT 0;
    SELECT COUNT(1) INTO payments_cnt FROM payments p WHERE p.merchant_uid = dup_merchant_uid COLLATE utf8_unicode_ci;

    SELECT payments_cnt, loop_cnt;
    IF loop_cnt > 0 THEN
        START TRANSACTION;
        WHILE i < loop_cnt AND i < 100 DO
            SET @imp_uid = CONCAT('imp_rody_will_delete_it-', ROUND(RAND() * 100000));
            SET @created_at = DATE_ADD('2023-09-07 16:58:47', INTERVAL FLOOR(48 * RAND()) HOUR);
            SET @status = IF(RAND() * 10 < 8, 'paid', 'cancelled');
            SET @cancelled_at = IF(@status = 'paid', NULL, DATE_ADD(@created_at, INTERVAL FLOOR(48 * RAND()) MINUTE));
            SET @paid_at = DATE_ADD(@created_at, INTERVAL FLOOR(48 * RAND()) MINUTE);
            SET @cancel_reason = IF(@status = 'paid', NULL, '관리자페이지취소');
            
            
            INSERT INTO payments (user_id, channel, currency, amount, vat, cancel_amount, name, imp_uid, merchant_uid, pg_tid,
                                  pg_provider, pg_id, pg_secret, cancel_password, pg_ext_key, pg_ext_priv, sandbox, pay_method,
                                  apply_num, card_code, card_quota, vbank_num, vbank_date, vbank_holder, vbank_code, is_escrow,
                                  buyer_name, buyer_email, buyer_tel, rev_buyer_tel, buyer_addr, buyer_postcode, custom_data,
                                  user_agent, user_ip, origin, app_scheme, m_redirect_url, status, paid_at, failed_at, cancelled_at,
                                  fail_reason, cancel_reason, confirm_url, notification_url, confirm_response, created, modified)
            VALUES (424, 'pc', 'KRW', 5000000, NULL, NULL, 'rody test', @imp_uid, dup_merchant_uid,
                    'someMID01012303231659163902', 'nice', 'someMID',
                    'dbW/d3tM72T5TrLeHAm1ph76MOmy1lFZSDH6KCTHlZrMcLHr3AJi9m04AOnv7Fh86ladh52h+wky+LMmP1PbLw==', '123456', NULL,
                    NULL, 0, 'card', '83450157', '365', 0, NULL, NULL, NULL, NULL, 0, 'Rody', 'rody@chai.finance', '010-1234-5678',
                    '87654321010', '서울시 성동구 성수이로 길20 16번지 JK타워 3층', '04787', NULL,
                    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36',
                    '172.25.0.1', 'http://sub_a.localhost:8001/index.html?env=local&version=1.2.0', NULL, NULL, @status,
                    @paid_at, NULL, @cancelled_at, NULL, @cancel_reason, @paid_at,
                    @cancelled_at, NULL, @created_at, @created_at);

            SET i = i + 1;
        END WHILE;
        COMMIT;
    END IF;

    SELECT COUNT(1) AS result_cnt FROM payments p WHERE p.merchant_uid = dup_merchant_uid;
    SELECT p.id FROM payments p WHERE p.merchant_uid = dup_merchant_uid;
END;

DROP PROCEDURE IF EXISTS insert_demo_payment_cancels_by_rody;

CREATE PROCEDURE insert_demo_payment_cancels_by_rody(IN payment_id INT, IN loop_cnt INT, IN cancel_amount INT)
BEGIN
    DECLARE payments_cnt INT;
    DECLARE i INT DEFAULT 0;
    SELECT COUNT(1) INTO payments_cnt FROM payments p WHERE p.id = payment_id;
    SELECT payments_cnt, loop_cnt;
    IF payments_cnt > 0 AND loop_cnt > 0 THEN
        START TRANSACTION;
        WHILE i < loop_cnt DO
            INSERT INTO payment_cancels (user_id,payment_id,pg_tid,cancel_amount,cancel_taxfree,reason,applied,created) VALUES
                (424,payment_id,'someMID01012303231659163902',cancel_amount,0,'관리자페이지취소','2023-03-23 17:09:03','2023-03-23 17:09:03');
            SET i = i + 1;
        END WHILE;
        COMMIT;
    END IF;

    SELECT COUNT(1) AS cancels_cnt FROM payment_cancels pc WHERE pc.payment_id = payment_id;
END;

DROP PROCEDURE IF EXISTS delete_demo_data_by_rody;

CREATE PROCEDURE delete_demo_data_by_rody()
BEGIN
    DECLARE done INT DEFAULT 0;
    DECLARE current_payment_id INT;
    DECLARE payment_id_cursor CURSOR FOR SELECT p.id FROM payments p WHERE p.merchant_uid = 'rody_large_transactions_test';
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1;
   
    START TRANSACTION;
    OPEN payment_id_cursor;
      
    read_loop: LOOP
        FETCH payment_id_cursor INTO current_payment_id;
        IF done THEN
          LEAVE read_loop;
        END IF;
    
        IF current_payment_id >= 370075616 THEN
            DELETE FROM payments WHERE id = current_payment_id;
            DELETE FROM payment_cancels WHERE payment_id = current_payment_id;
            DELETE FROM payment_notices WHERE payment_id = current_payment_id;
        END IF;
    END LOOP;
    
    CLOSE payment_id_cursor;

    COMMIT;
END;


SHOW PROCEDURE STATUS;

DROP PROCEDURE IF EXISTS insert_demo_payments_by_rody;
DROP PROCEDURE IF EXISTS delete_demo_data_by_rody;
DROP PROCEDURE IF EXISTS insert_demo_payment_cancels_by_rody;
DROP PROCEDURE IF EXISTS insert_demo_payment_notices_by_rody;
```
