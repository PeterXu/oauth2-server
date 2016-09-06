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


### /checktoken

```
    Method:
        POST
    Params:
        access_token
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

