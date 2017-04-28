-- DELETE FROM mysql.user WHERE user='todouser'; FLUSH PRIVILEGES;
CREATE USER 'todouser'@'localhost' IDENTIFIED BY 'todopswd'; FLUSH PRIVILEGES;
GRANT ALL PRIVILEGES ON *.* TO 'todouser'@'localhost'; FLUSH PRIVILEGES;
CREATE USER 'todouser'@'%' IDENTIFIED BY 'todopswd'; FLUSH PRIVILEGES;
GRANT ALL PRIVILEGES ON *.* TO 'todouser'@'%'; FLUSH PRIVILEGES;

CREATE DATABASE todolists;
USE todolists;

-- DROP TABLE IF EXISTS `items`;
-- DROP TABLE IF EXISTS `lists`;
-- DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
       `id` BIGINT NOT NULL AUTO_INCREMENT,
       `username` varchar(255) UNIQUE NOT NULL,
       `password` varchar(255) NOT NULL DEFAULT "",
       `sessionid` varchar(255),
       PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `lists` (
       `id` BIGINT NOT NULL AUTO_INCREMENT,
       `userid` BIGINT NOT NULL,
       `title` VARCHAR(255) NOT NULL DEFAULT "",
       `priority` INT NOT NULL DEFAULT 0,
       `created_at` BIGINT  NOT NULL DEFAULT 0,
       PRIMARY KEY (`id`),
       UNIQUE KEY (`userid`, `title`),
       FOREIGN KEY(`userid`) REFERENCES users(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `items` (
       `id` BIGINT NOT NULL AUTO_INCREMENT,
       `userid` BIGINT NOT NULL,
       `listid` BIGINT NOT NULL,
       `title` VARCHAR(255) NOT NULL DEFAULT "",
       `item` TEXT NOT NULL,
       `priority` INT NOT NULL DEFAULT 0,
       `created_at` BIGINT NOT NULL DEFAULT 0,
       `due_date` BIGINT NOT NULL DEFAULT 0,
       PRIMARY KEY(`id`, `userid`, `listid`),
       FOREIGN KEY(`userid`) REFERENCES users(`id`),
       FOREIGN KEY(`listid`) REFERENCES lists(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

