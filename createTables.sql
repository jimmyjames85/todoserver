-- DELETE FROM mysql.user WHERE user='todouser'; FLUSH PRIVILEGES;
CREATE USER 'todouser'@'localhost' IDENTIFIED BY 'todopswd'; FLUSH PRIVILEGES;
GRANT ALL PRIVILEGES ON *.* TO 'todouser'@'localhost'; FLUSH PRIVILEGES;
CREATE USER 'todouser'@'%' IDENTIFIED BY 'todopswd'; FLUSH PRIVILEGES;
GRANT ALL PRIVILEGES ON *.* TO 'todouser'@'%'; FLUSH PRIVILEGES;

CREATE DATABASE todolists;
USE todolists;

DROP TABLE IF EXISTS `items`;
DROP TABLE IF EXISTS `lists`;
DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
       `id` BIGINT UNIQUE NOT NULL AUTO_INCREMENT,
       `username` varchar(255) UNIQUE NOT NULL,
       `password` varchar(255) NOT NULL DEFAULT "",
       `apikey`   varchar(255) UNIQUE,
       `sessionid` varchar(255) UNIQUE,
       PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `lists` (
       `id` BIGINT NOT NULL AUTO_INCREMENT,
       `userid` BIGINT NOT NULL,
       `title` VARCHAR(255) NOT NULL DEFAULT "",
       `priority` INT NOT NULL DEFAULT 0,
       `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
       PRIMARY KEY (`id`),
       UNIQUE KEY (`userid`, `title`),
       FOREIGN KEY(`userid`) REFERENCES users(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `items` (
       `id` BIGINT NOT NULL UNIQUE AUTO_INCREMENT,
       `userid` BIGINT NOT NULL,
       `listid` BIGINT NOT NULL,
       `item` VARCHAR(255) NOT NULL DEFAULT "",
       `details` TEXT NOT NULL,
       `priority` INT NOT NULL DEFAULT 0,
       `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,       
       `due_date` TIMESTAMP,
       PRIMARY KEY(`id`, `userid`),
       UNIQUE KEY (`userid`, `listid`, `item`),
       FOREIGN KEY(`userid`) REFERENCES users(`id`),
       FOREIGN KEY(`listid`) REFERENCES lists(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO users (username, password) VALUES ('todo', 'todo');
INSERT INTO lists (userid, title)  VALUES (1, 'grocery'),  (1, 'work'),  (1, 'home');
INSERT INTO items (userid, listid, item) VALUES (1, 1, 'milk'), (1, 1, 'oj'), (1, 1, 'bread');
INSERT INTO items (userid, listid, item) VALUES (1, 2, 'standup'), (1, 2, 'story 3'), (1, 2, 'ping pong');
INSERT INTO items (userid, listid, item) VALUES (1, 3, 'clean toilet'), (1, 3, 'wash dog'), (1, 3, 'clean gun');

