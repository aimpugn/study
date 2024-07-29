-- https://hyorock.tistory.com/52
-- 전체 database character set 확인
SELECT
    SCHEMA_NAME AS 'database',
    DEFAULT_CHARACTER_SET_NAME AS 'character_set',
    DEFAULT_COLLATION_NAME AS 'collation'
FROM information_schema.SCHEMATA;

-- 하나의 database character set 확인
USE database_name;
SHOW VARIABLES LIKE 'character_set_database';

-- table collation 확인
SHOW TABLE STATUS WHERE NAME LIKE 'table_name';

-- column collation 확인
SHOW FULL COLUMNS FROM table_name;

-- charset, collation 조회
SHOW VARIABLES
    WHERE VARIABLE_NAME LIKE '%coll%'
        OR VARIABLE_NAME LIKE '%char%'
        OR VARIABLE_NAME ='init_connect';