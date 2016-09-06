misc docs
=========

password & random
-----------------

htpasswd use bcrypt encryption
```
htpasswd -nb -B testuser test123456
```

generate random 64bytes secret
```
pwgen -Bsv1 64
```



The call flow in oauth2.0 lib
-----------------------------

### flow of token  
    -> ClientInfoHandler -> PasswordAuthorizationHandler  
    -> ClientAuthorizedHandler -> ClientScopeHandler  
```
/token? 
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
            => Server.CheckGrantType()              for valid 'grant_type'
            => my.ClientAuthorizedHandler(..)       for clientID's 'grant_type'  
            => my.ClientScopeHandler(..)            for clientID's 'scope'  
            => Manager.GenerateAccessToken(..)  
                => Manager.GetClient(..) clientID  
                    => check from oauth2.ClientStore  
                    => oauth2.ClientStore.GetByID()  
                    => return Client  
                => check clientSecret != client.GetSecret()  
                => GenerateBasic Client/userID/CreateAt  
                => TokenInfo clientID/userID/redirectURI/scope  
                => return accessToken  
```
 

### flow of authorize - UserAuthorizationHandler
```
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
```


