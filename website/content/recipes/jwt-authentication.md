---
title: JWT Authentication
menu:
  side:
    parent: recipes
    weight: 11
---

Most applications dealing with client authentication will require a more secure
mechanism than that provided by [basic authentication](https://github.com/labstack/echo/blob/master/middleware/auth.go). [JSON Web Tokens](http://jwt.io/)
are one such mechanism - JWTs are a compact means of transferring cryptographically
signed claims between the client and server.

This recipe demonstrates the use of a simple JWT authentication Echo middleware
using Dave Grijalva's [jwt-go](https://github.com/dgrijalva/jwt-go). This middleware
expects the token to be present in an Authorization HTTP header using the method
"Bearer", although JWTs are also frequently sent using cookies, the request URL,
or even the request body. We will use the HS236 signing method, note that several
other algorithms are available.

`server.go`

{{< embed "jwt-authentication/server.go" >}}

Run `server.go` and making a request to the root path `/` returns a 200 OK response,
as this route does not use our JWT authentication middleware. Sending requests to
`/restricted` (our authenticated route) with either no Authorization header or invalid
Authorization headers / tokens will return 401 Unauthorized.

```sh
# Unauthenticated route
$ curl localhost:1323/  => No auth required for this route.

# No Authentication header
$ curl localhost:1323/restricted  => Unauthorized

# Invalid Authentication method
$  curl localhost:1323/restricted -H "Authorization: Invalid " => Unauthorized

# Invalid token
$  curl localhost:1323/restricted -H "Authorization: Bearer InvalidToken" => Unauthorized
```

Running `token.go` (source) will print JWT that is valid against this middleware
to stdout. You can use this token to test succesful authentication on the `/restricted` path.

{{< embed "jwt-authentication/token/token.go" >}}

```sh
# Valid token
$  curl localhost:1323/restricted -H "Authorization: Bearer <token>" => Access granted with JWT.
```

### Maintainers

- [axdg](https://github.com/axdg)

### [Source Code](https://github.com/vishr/recipes/blob/master/echo/recipes/jwt-authentication)
