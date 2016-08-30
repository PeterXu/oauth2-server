
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

call flow:
    ClientInfoHandler -> PasswordAuthorizationHandler -> ClientAuthorizedHandler -> ClientScopeHandler


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
<code>
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
</code>

<code>
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
</code>


Test
----

### username/password
<code>
htpasswd -nb -B testuser test123456 # use bcrypt encryption
curl -XPOST "localhost:6543/token?grant_type=password&username=testuser&password=test123456"
curl -XPOST "localhost:6543/token?grant_type=password&username=testuser&password=test123456&scope=app"
</code>


### Authorization Code Grant
response_type - required(code)  
client_id - required  
redirect_uri - optional  
scope - optional  
state - optional  

The authorization code grant type is used to obtain both access
   tokens and refresh tokens and is optimized for confidential clients.

it is similar to implicit (access token request)

<code>
GET /authorize?response_type=code&client_id=s6BhdRkqt3&state=xyz&scope=app
&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
Host: server.example.com
</code>

<code>
HTTP/1.1 302 Found
Location: https://client.example.com/cb?code=SplxlOBeZQQYbYS6WxSbIA&state=xyz

HTTP/1.1 302 Found
Location: https://client.example.com/cb?error=access_denied&state=xyz
</code>


### Access Token Request

##### 1. Access Token Request - authorization_code
grant_type - required(authorization_code)  
code - required  
redirect_uri - required  
client_id - required  

<code>
POST /token HTTP/1.1
Host: server.example.com
Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&code=SplxlOBeZQQYbYS6WxSbIA
&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb
</code>

<code>
HTTP/1.1 200 OK
Content-Type: application/json;charset=UTF-8
Cache-Control: no-store
Pragma: no-cache

{
"access_token":"2YotnFZFEjr1zCsicMWpAA",
"token_type":"example",
"expires_in":3600,
"refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA",
"example_parameter":"example_value"
}
</code>

##### 2. Access Token Request - implicit
The implicit grant type is used to obtain access tokens (it does not
support the issuance of refresh tokens) and is optimized for public
clients known to operate a particular redirection URI(e.g. browser).
The default token_type is "Bearer";

response_type - required(token)  
client_id - required  
redirect_uri - optional  
scope - optional  
state - RECOMMENDED, maintain state by client  

<code>
GET /authorize?response_type=token&client_id=s6BhdRkqt3&state=xyz
&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
Host: server.example.com
</code>

<code>
HTTP/1.1 302 Found
Location: http://example.com/cb#access_token=2YotnFZFEjr1zCsicMWpAA
&state=xyz&token_type=example&expires_in=3600

HTTP/1.1 302 Found
Location: https://client.example.com/cb#error=access_denied&state=xyz
</code>


##### 3. Access Token Request - password grant
grant_type - required(password)  
username - required  
password - required  
scope - optional  

<code>
POST /token HTTP/1.1
Host: server.example.com
Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
Content-Type: application/x-www-form-urlencoded

grant_type=password&username=johndoe&password=A3ddj3w
</code>

<code>
HTTP/1.1 200 OK
Content-Type: application/json;charset=UTF-8
Cache-Control: no-store
Pragma: no-cache

{
"access_token":"2YotnFZFEjr1zCsicMWpAA",
"token_type":"example",
"expires_in":3600,
"refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA",
"example_parameter":"example_value"
}
</code>


##### 4. Access Token Request - client_credentials grant
grant_type - required(client_credentials)  
scope - optional  

(no additional authorization request is needed, for authorized clients.)

<code>
POST /token HTTP/1.1
Host: server.example.com
Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
</code>

<code>
HTTP/1.1 200 OK
Content-Type: application/json;charset=UTF-8
Cache-Control: no-store
Pragma: no-cache

{
"access_token":"2YotnFZFEjr1zCsicMWpAA",
"token_type":"example",
"expires_in":3600,
"example_parameter":"example_value"
}
</code>

##### 5. Access Token Request - refresh_token grant
grant_type - required(refresh_token)  
refresh_token - required
client_id - 
client_secret - 

<code>
POST /token HTTP/1.1
Host: server.example.com
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&refresh_token=tGzv3JOkF0XG5Qx2TlKWIA
&client_id=s6BhdRkqt3&client_secret=7Fjfp0ZBr1KtDRbnfVdmIw
</code>



