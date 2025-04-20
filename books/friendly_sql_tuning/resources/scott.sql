-- 친절한 SQL 튜닝 책 실습 데이터 실습 데이터를 MySQL로 변환한 스크립트입니다.
-- Ref:
-- - https://cafe.naver.com/f-e/cafes/28788317/articles/4661
CREATE TABLE DEPT (
  -- DEPTNO NUMBER (2) CONSTRAINT PK_DEPT PRIMARY KEY,
  -- TINYINT(0 ∼ 255) 또는 SMALLINT(−32,768 ∼ 32,767)로 대신합니다.
  DEPTNO TINYINT UNSIGNED NOT NULL,
  -- VARCHAR2 -> VARCHAR
  DNAME VARCHAR(14),
  LOC VARCHAR(13),
  CONSTRAINT PK_DEPT PRIMARY KEY (DEPTNO)
) ENGINE = InnoDB;


INSERT INTO
  DEPT
VALUES
  (10, 'ACCOUNTING', 'NEW YORK');


INSERT INTO
  DEPT
VALUES
  (20, 'RESEARCH', 'DALLAS');


INSERT INTO
  DEPT
VALUES
  (30, 'SALES', 'CHICAGO');


INSERT INTO
  DEPT
VALUES
  (40, 'OPERATIONS', 'BOSTON');


COMMIT;


CREATE TABLE EMP (
  -- EMPNO NUMBER (4) CONSTRAINT PK_EMP PRIMARY KEY,
  -- 1. NUMBER(4) 는 −9,999 ∼ 9,999 또는 0 ∼ 9,999(UNSIGNED) 범위의 정수형이므로,
  --    SMALLINT(−32,768 ∼ 32,767)로 대신합니다.
  -- 2. 컬럼 수준 제약에 이름을 붙일 수 있는 구문을 지원하지 않습니다.
  --    컬럼 바로 뒤에는 PRIMARY KEY만 쓰거나, 제약 이름을 붙이려면 테이블 수준으로 따로 선언해야 합니다.
  EMPNO SMALLINT UNSIGNED PRIMARY KEY,
  ENAME VARCHAR(10),
  JOB VARCHAR(9),
  MGR SMALLINT UNSIGNED,
  HIREDATE DATE,
  SAL DECIMAL(7, 2),
  COMM DECIMAL(7, 2),
  -- DEPTNO NUMBER(2)
  DEPTNO TINYINT UNSIGNED
);


INSERT INTO
  EMP
VALUES
  (
    7369,
    'SMITH',
    'CLERK',
    7902,
    STR_TO_DATE ('17-12-1980', '%e-%c-%Y'),
    800,
    NULL,
    20
  );


INSERT INTO
  EMP
VALUES
  (
    7499,
    'ALLEN',
    'SALESMAN',
    7698,
    STR_TO_DATE ('20-2-1981', '%e-%c-%Y'),
    1600,
    300,
    30
  );


INSERT INTO
  EMP
VALUES
  (
    7521,
    'WARD',
    'SALESMAN',
    7698,
    STR_TO_DATE ('22-2-1981', '%e-%c-%Y'),
    1250,
    500,
    30
  );


INSERT INTO
  EMP
VALUES
  (
    7566,
    'JONES',
    'MANAGER',
    7839,
    STR_TO_DATE ('2-4-1981', '%e-%c-%Y'),
    2975,
    NULL,
    20
  );


INSERT INTO
  EMP
VALUES
  (
    7654,
    'MARTIN',
    'SALESMAN',
    7698,
    STR_TO_DATE ('28-9-1981', '%e-%c-%Y'),
    1250,
    1400,
    30
  );


INSERT INTO
  EMP
VALUES
  (
    7698,
    'BLAKE',
    'MANAGER',
    7839,
    STR_TO_DATE ('1-5-1981', '%e-%c-%Y'),
    2850,
    NULL,
    30
  );


INSERT INTO
  EMP
VALUES
  (
    7782,
    'CLARK',
    'MANAGER',
    7839,
    STR_TO_DATE ('9-6-1981', '%e-%c-%Y'),
    2450,
    NULL,
    10
  );


INSERT INTO
  EMP
VALUES
  (
    7788,
    'SCOTT',
    'ANALYST',
    7566,
    STR_TO_DATE ('13-7-1987', '%e-%c-%Y') -85,
    3000,
    NULL,
    20
  );


INSERT INTO
  EMP
VALUES
  (
    7839,
    'KING',
    'PRESIDENT',
    NULL,
    STR_TO_DATE ('17-11-1981', '%e-%c-%Y'),
    5000,
    NULL,
    10
  );


INSERT INTO
  EMP
VALUES
  (
    7844,
    'TURNER',
    'SALESMAN',
    7698,
    STR_TO_DATE ('8-9-1981', '%e-%c-%Y'),
    1500,
    0,
    30
  );


INSERT INTO
  EMP
VALUES
  (
    7876,
    'ADAMS',
    'CLERK',
    7788,
    STR_TO_DATE ('13-7-1987', '%e-%c-%Y') -51,
    1100,
    NULL,
    20
  );


INSERT INTO
  EMP
VALUES
  (
    7900,
    'JAMES',
    'CLERK',
    7698,
    STR_TO_DATE ('3-12-1981', '%e-%c-%Y'),
    950,
    NULL,
    30
  );


INSERT INTO
  EMP
VALUES
  (
    7902,
    'FORD',
    'ANALYST',
    7566,
    STR_TO_DATE ('3-12-1981', '%e-%c-%Y'),
    3000,
    NULL,
    20
  );


INSERT INTO
  EMP
VALUES
  (
    7934,
    'MILLER',
    'CLERK',
    7782,
    STR_TO_DATE ('23-1-1982', '%e-%c-%Y'),
    1300,
    NULL,
    10
  );


COMMIT;
