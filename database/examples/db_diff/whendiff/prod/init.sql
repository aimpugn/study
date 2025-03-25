GRANT ALL PRIVILEGES
    ON some_service.*
    TO 'prod_user'@'%';
FLUSH PRIVILEGES;

CREATE TABLE users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(50) NOT NULL UNIQUE,
  password_hash VARCHAR(256) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  title VARCHAR(200),
  body TEXT,
  only_at_prod VARCHAR(50), -- dev DB에 없는 컬럼
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE comments ( -- prod DB에 없는 테이블
  id INT AUTO_INCREMENT PRIMARY KEY,
  post_id INT NOT NULL,
  comment TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (post_id) REFERENCES posts (id)
);

INSERT INTO
  users (username, password_hash)
VALUES
  ('devuser1', 'hash1'),
  ('devuser2', 'hash2');


INSERT INTO
  posts (user_id, title, body)
VALUES
  (1, 'Dev Title 1', 'Dev Body 1'),
  (2, 'Dev Title 2', 'Dev Body 2');
