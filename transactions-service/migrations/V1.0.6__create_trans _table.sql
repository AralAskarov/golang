CREATE TABLE transactions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    uuid BINARY(16) NOT NULL COMMENT 'users.uuid',
    amount INT NOT NULL,
    type ENUM('deposit', 'withdrawal') NOT NULL,
    time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);