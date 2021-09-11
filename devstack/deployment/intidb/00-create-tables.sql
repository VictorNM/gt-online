CREATE TABLE IF NOT EXISTS `users`
(
    `email`      varchar(255) NOT NULL,
    `password`   varchar(255) NOT NULL,
    `first_name` varchar(255) NOT NULL,
    `last_name`  varchar(255) NOT NULL,
    PRIMARY KEY (`email`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE IF NOT EXISTS `admin_users`
(
    `email`      varchar(255) NOT NULL,
    `last_login` datetime     NULL,
    PRIMARY KEY (`email`),
    FOREIGN KEY (email) REFERENCES users (email) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE IF NOT EXISTS `regular_users`
(
    `email`        varchar(255) NOT NULL,
    `birthdate`    date         NULL,
    `sex`          char(1)      NULL,
    `current_city` varchar(50)  NULL,
    `hometown`     varchar(50)  NULL,
    PRIMARY KEY (`email`),
    FOREIGN KEY (email) REFERENCES users (email) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE IF NOT EXISTS `interests`
(
    `email`    varchar(255) NOT NULL,
    `interest` varchar(50)  NOT NULL,
    PRIMARY KEY (`email`, `interest`),
    FOREIGN KEY (email) REFERENCES regular_users (email) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE IF NOT EXISTS `school_types`
(
    `type_name` varchar(50) NOT NULL,
    PRIMARY KEY (`type_name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE IF NOT EXISTS `schools`
(
    `school_name` varchar(255) NOT NULL,
    `type`        varchar(50)  NOT NULL,
    PRIMARY KEY (`school_name`),
    FOREIGN KEY (type) REFERENCES school_types (type_name) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE IF NOT EXISTS `attends`
(
    `email`          varchar(255) NOT NULL,
    `school_name`    varchar(50)  NOT NULL,
    `year_graduated` int          NULL,
    UNIQUE (`email`, `school_name`, `year_graduated`),
    FOREIGN KEY (email) REFERENCES regular_users (email) ON DELETE CASCADE,
    FOREIGN KEY (school_name) REFERENCES schools (school_name)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE IF NOT EXISTS `employers`
(
    `employer_name` varchar(50) NOT NULL,
    PRIMARY KEY (`employer_name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE IF NOT EXISTS `employments`
(
    `email`         varchar(255) NOT NULL,
    `employer_name` varchar(50)  NOT NULL,
    `job_title`     varchar(50)  NOT NULL,
    UNIQUE (`email`, `employer_name`, `job_title`),
    FOREIGN KEY (email) REFERENCES regular_users (email) ON DELETE CASCADE,
    FOREIGN KEY (employer_name) REFERENCES employers (employer_name)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE IF NOT EXISTS `friendships`
(
    `email`          varchar(255) NOT NULL,
    `friend_email`   varchar(255) NOT NULL,
    `relationship`   varchar(50)  NOT NULL,
    `date_connected` datetime     NOT NULL,
    PRIMARY KEY (`email`, `friend_email`),
    FOREIGN KEY (email) REFERENCES regular_users (email) ON DELETE CASCADE,
    FOREIGN KEY (friend_email) REFERENCES regular_users (email) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;
