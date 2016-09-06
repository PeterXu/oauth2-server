DB structure
============


mysql
-----

insert user with password
```
insert into users values(
    '', uuid(), 'testuser', '', '', hash_password, '0'
)
```

check user's password
```
select substring(password,0,64) as salt, substring(password,64) as sha_passwd
    from user where user.username = username;
sha(user_passwd, salt) ==? password
```


create mysql database
```
CREATE DATABASE IF NOT EXISTS oauth charset=utf8;
USE oauth;
CREATE TABLE `users` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `uid` varchar(32) NOT NULL UNIQUE,
    `password` varchar(128) NOT NULL,

    `username` varchar(64) DEFAULT NULL UNIQUE,
    `email` varchar(64) DEFAULT NULL UNIQUE,
    `cell` varchar(64) DEFAULT NULL UNIQUE,

    `reset_password` tinyint(4) NOT NULL DEFAULT 0,
    `retry_count` tinyint(4) NOT NULL DEFAULT 0,
    `status` enum('active','deleted','inactive') NOT NULL DEFAULT 'active',

    `login_count` int(11) NOT NULL DEFAULT 0,
    `last_login` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
```

