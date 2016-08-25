
The flow of /token? 
------------------
/token? =>
    Server.HandleTokenRequest(..) => 
        => Server.ValidationTokenRequest(..)
            => only POST not GET
            => require <grant_type>
                => authorization_code/password/client_credentials/refresh_token
            => my.ClientInfoHandler(..)
                => return 'clientID/clientSecret' for token generation
                => which should be exist in oauth2.ClientStorage(manager.MapClientStorage)
            => option <scope> 
            => my.PasswordAuthorizationHandler()
                => check <username>/<password>
                => htpasswd/mysql/..
                => return userID
            => TokenGenerateRequest(PasswordCredential):
                => clientID/clientSecret/Scope/UserID
        => Server.GetAccessToken(..) password/TokenGenerateRequest
            => Server.CheckGrantType()              for valid <grant_type>
            => my.ClientAuthorizedHandler(..)       for clientID's <grant_type>
            => my.ClientScopeHandler(..)            for clientID's <scope>
            => Manager.GenerateAccessToken(..)
                => Manager.GetClient(..) clientID
                    => check from oauth2.ClientStore
                    => oauth2.ClientStore.GetByID()
                    => return Client
                => check clientSecret != client.GetSecret()
                => GenerateBasic Client/userID/CreateAt
                => TokenInfo clientID/userID/redirectURI/scope
                => return accessToken


/authorize? =>
    Server.HandleAuthorizeRequest(..) =>
        => Server.ValidationAuthorizeRequest(..)
            => require <redirect_uri>
            => require <client_id>
            => require <response_type>  => code/token
            => option <state>
            => option <scope>
        => my.UserAuthorizationHandler(..)
            => return userID
        => my.AuthorizeScopeHandler(..)
            => return scope
        => my.AccessTokenExpHandler(..)
            => return AccessTokenExp
        => Server.GetAuthorizeToken()
            => my.ClientAuthorizedHandler(..)       for clientID's <grant_type>
            => my.ClientScopeHandler(..)            for clientID's <scope>
            => Manager.GenerateAuthToken()
        => Server.GetAuthorizeData()
            => Server.GetTokenData()


mysql
-----
```
pwgen -Bsv1 64
username
user_passwd=sha(plain_passwd);

# generate
salt = random64char()
sha_passwd=sha(user_passwd, salt)
insert into users values(
    '', uuid(), 'testuser', '', '', contact($salt, $sha_passwd), '0'
)

# verify
select substring(password,0,64) as salt, substring(password,64) as sha_passwd  
    from user where user.username = username;
sha(user_passwd, salt) ==? password
```

```
USE oauth;
CREATE TABLE `users` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `uid` varchar(32) NOT NULL UNIQUE,
    `username` varchar(64) NOT NULL UNIQUE,
    `email` varchar(64) DEFAULT NULL UNIQUE,
    `cell` varchar(64) DEFAULT NULL UNIQUE,

    `password` varchar(128) NOT NULL,
    `reset_password` tinyint(4) NOT NULL DEFAULT 0,
    `retry_count` tinyint(4) NOT NULL DEFAULT 0,
    `status` enum('active','deleted','inactive') NOT NULL DEFAULT 'active',

    `login_count` int(11) NOT NULL DEFAULT 0,
    `last_login` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
```


Test
----

```
htpasswd -nb -B testuser test123456 # use bcrypt encryption
curl -XPOST "localhost:6543/token?grant_type=password&username=testuser&password=test123456"
curl -XPOST "localhost:6543/token?grant_type=password&username=testuser&password=test123456&scope=app"
```

```
gClientID="76dbb7ac-6da8-11e6-84c6-1b976b623e41"
redirectURI="http://localhost"
curl -XGET  "localhost:6543/authorize?client_id=$gClientID&response_type=code&redirect_uri=$redirectURI&state=good&scope=app"
```

