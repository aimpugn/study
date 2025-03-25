GRANT ALL PRIVILEGES
    ON some_service.*
    TO 'dev_user'@'%';
FLUSH PRIVILEGES;

CREATE TABLE users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(50) NOT NULL UNIQUE,
  password_hash VARCHAR(256) NOT NULL,
  email VARCHAR(100), -- prod DB에 없는 컬럼
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE posts (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  title VARCHAR(200),
  body TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, -- prod DB에 없는 컬럼
  FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE comments (
  id INT AUTO_INCREMENT PRIMARY KEY,
  post_id INT NOT NULL,
  comment TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- prod DB에 없는 칼럼
  FOREIGN KEY (post_id) REFERENCES posts (id)
);

CREATE TABLE only_at_dev ( -- prod DB에 없는 테이블
  id INT AUTO_INCREMENT PRIMARY KEY,
  ref_id INT NOT NULL,
  extra TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO
  users (username, password_hash, email)
VALUES
  ('produser1', 'hash1', 'user1@example.com'),
  ('produser2', 'hash2', 'user2@example.com');


INSERT INTO
  posts (user_id, title, body)
VALUES
  (1, 'Prod Title 1', 'Prod Body 1'),
  (2, 'Prod Title 2', 'Prod Body 2');


INSERT INTO
  comments (post_id, comment)
VALUES
  (1, 'Nice post!'),
  (2, 'Thanks for sharing.');
