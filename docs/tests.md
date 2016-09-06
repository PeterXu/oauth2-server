Test
====


how to use API
--------------


### get token with username/password
```
curl -XPOST "localhost:6543/token?grant_type=password&username=testuser&password=test123456"
curl -XPOST "localhost:6543/token?grant_type=password&username=testuser&password=test123456&scope=app"
```


### Authorization Code Grant
>response_type - required(code)  
client_id - required  
redirect_uri - optional  
scope - optional  
state - optional  

The authorization code grant type is used to obtain both access
   tokens and refresh tokens and is optimized for confidential clients.

it is similar to implicit (access token request)

```
GET /authorize?response_type=code&client_id=s6BhdRkqt3&state=xyz&scope=app
&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
Host: server.example.com

HTTP/1.1 302 Found
Location: https://client.example.com/cb?code=SplxlOBeZQQYbYS6WxSbIA&state=xyz

HTTP/1.1 302 Found
Location: https://client.example.com/cb?error=access_denied&state=xyz
```

Note: the client_id and redirect_uri's root should be the same as configure in oauth2-server.
here we have a default client_id if not in request-uri.
```
http://localhost:6543/authorize?response_type=code&client_id=osso1&state=xyz&scope=app&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb
```

### Access Token Request

##### 1. Access Token Request - authorization_code
>grant_type - required(authorization_code)  
code - required  
redirect_uri - required  
client_id - required  

request
```
POST /token HTTP/1.1
Host: server.example.com
Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&code=SplxlOBeZQQYbYS6WxSbIA
&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb
```

response
```
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
```

##### 2. Access Token Request - implicit
The implicit grant type is used to obtain access tokens (it does not
support the issuance of refresh tokens) and is optimized for public
clients known to operate a particular redirection URI(e.g. browser).
The default token_type is "Bearer";

>response_type - required(token)  
client_id - required  
redirect_uri - optional  
scope - optional  
state - RECOMMENDED, maintain state by client  

request
```
GET /authorize?response_type=token&client_id=s6BhdRkqt3&state=xyz
&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
Host: server.example.com
```

response
```
HTTP/1.1 302 Found
Location: http://example.com/cb#access_token=2YotnFZFEjr1zCsicMWpAA
&state=xyz&token_type=example&expires_in=3600

HTTP/1.1 302 Found
Location: https://client.example.com/cb#error=access_denied&state=xyz
```


##### 3. Access Token Request - password grant
>grant_type - required(password)  
username - required  
password - required  
scope - optional  

request
```
POST /token HTTP/1.1
Host: server.example.com
Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
Content-Type: application/x-www-form-urlencoded

grant_type=password&username=johndoe&password=A3ddj3w
```

response
```
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
```


##### 4. Access Token Request - client_credentials grant
>grant_type - required(client_credentials)  
scope - optional  

(no additional authorization request is needed, for authorized clients.)

request
```
POST /token HTTP/1.1
Host: server.example.com
Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
```

response
```
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
```

##### 5. Access Token Request - refresh_token grant
>grant_type - required(refresh_token)  
refresh_token - required  
client_id - required but now with default if no 
client_secret - required but now with default if no

```
POST /token HTTP/1.1
Host: server.example.com
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&refresh_token=tGzv3JOkF0XG5Qx2TlKWIA
&client_id=s6BhdRkqt3&client_secret=7Fjfp0ZBr1KtDRbnfVdmIw
```


Test cases
----------

```

5.
curl -XPOST "localhost:6543/token?grant_type=password&username=testuser4&password=testpass4&scope=app"
curl -XPOST "localhost:6543/token?grant_type=refresh_token&refresh_token=last_access_token"
curl -XPOST "localhost:6543/token?grant_type=refresh_token&refresh_token=last_refresh_token"

4.
curl -XPOST "localhost:6543/token?grant_type=client_credentials&scope=app"

3.
curl -XPOST "localhost:6543/token?grant_type=password&username=testuser4&password=testpass4&scope=app"

2.for browser

1.
curl -XGET "localhost:6543/token?grant_type=authorization_code&code=SplxlOBeZQQYbYS6WxSbIA&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb"

0.
curl -XGET "localhost:6543/authorize?response_type=code&client_id=s6BhdRkqt3&state=xyz&scope=app&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb"

-1. for code
curl -v -XPOST "localhost:6543/code?username=testuser4&password=testpass4"


```

