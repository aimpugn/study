CREATE DATABASE IF NOT EXISTS friendly;

USE friendly;

-- 사용자 및 권한
-- '%'는 원격 호스트(외부 IP, Docker host 등)에서 접속한 경우입니다.
-- CREATE USER IF NOT EXISTS 'rody'@'%' IDENTIFIED BY 'rody';
CREATE USER IF NOT EXISTS 'rody'@'localhost' IDENTIFIED BY 'rody';
GRANT ALL PRIVILEGES
    ON friendly.*
--     TO 'rody'@'%';
    TO 'rody'@'localhost';
FLUSH PRIVILEGES;


CREATE TABLE IF NOT EXISTS employees (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100),
  department VARCHAR(50),
  salary INT
);


INSERT INTO
  employees (name, department, salary)
VALUES
  ('Alice', 'HR', 5000),
  ('Bob', 'IT', 7000),
  ('Charlie', 'Sales', 6000);
