CREATE TABLE IF NOT EXISTS `users`
(
    `email`      varchar(50) NOT NULL,
    `password`   varchar(50) NOT NULL,
    `first_name` varchar(50) NOT NULL,
    `last_name`  varchar(50) NOT NULL,
    PRIMARY KEY (`email`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;