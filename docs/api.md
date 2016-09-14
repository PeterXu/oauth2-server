REST API
========

### /token
```
    Method:
        POST
    Params:
        grant_type
            Required. Value must be set to password
        username
            Required. The resource owner username, encoded as UTF-8.
        password
            Required. The resource owner password, encoded as UTF-8.
        scope
            Optional. The scope of the access request.
    Content-Type:
        application/x-www-form-urlencoded
    Response:
        HTTP/1.1 200 OK
        Content-Type: application/json;charset=UTF-8
        Cache-Control: no-store
        Pragma: no-cache

        { "access_token":"Qwe1235rwersdgasdfghjkyuiyuihfgh",
        "token_type":"bearer",
        "expires_in":3600,
        "scope": "exampleScope" }
```


### /check
```
    Method:
        POST
    Params:
        access_token|refresh_token
            Required. Value of the token to be checked
        username
            Required. The resource owner username, encoded as UTF-8.
        scope
            Optional. The scope of the access request.
    Content-Type:
        application/x-www-form-urlencoded
    Response:
        If successful: HTTP/1.1 200 OK
```

### /signup
```
    Method:
        POST
    Params:
        username
            Required. The resource owner username, encoded as UTF-8.
        password1
            Required. The username's password
        password2
            Required. The repeated password
    Content-Type:
        application/x-www-form-urlencoded
    Response:
        If successful: HTTP/1.1 200 OK
```

### /reset
```
    Method:
        POST
    Params:
        username
            Required. The resource owner username, encoded as UTF-8.
        password
            Required. The password for owner username
        password1
            Required. The new password
        password2
            Required. The repeated new password
    Content-Type:
        application/x-www-form-urlencoded
    Response:
        If successful: HTTP/1.1 200 OK
```


Web Browser
===========

```
http://localhost:6543/signin

http://localhost:6543/signup

http://localhost:6543/reset

http://localhost:6543/authorize?response_type=code&client_id=osso1&state=xyz&scope=app&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb

```

