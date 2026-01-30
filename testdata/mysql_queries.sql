-- sample SQL for tests
CREATE TABLE users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255),
  tags JSON
);

SELECT GROUP_CONCAT(name) FROM users;
