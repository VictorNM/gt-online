CREATE TABLE IF NOT EXISTS `users`
(
    `email`      varchar(255) NOT NULL,
    `password`   varchar(255) NOT NULL,
    `first_name` varchar(255) NOT NULL,
    `last_name`  varchar(255) NOT NULL,
    PRIMARY KEY (`email`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;