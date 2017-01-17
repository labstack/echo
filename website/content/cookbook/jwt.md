+++
title = "JWT"
description = "JWT example for Echo"
[menu.main]
  name = "JWT"
  identifier = "example-jwt"
  parent = "cookbook"
  weight = 11
+++

- JWT authentication using HS256 algorithm.
- JWT is retrieved from `Authorization` request header.

## Server using Map claims

`server.go`

{{< embed "jwt/map-claims/server.go" >}}

## Server using custom claims

`server.go`

{{< embed "jwt/custom-claims/server.go" >}}

## Client

`curl`

### Login

Login using username and password to retrieve a token.

```sh
curl -X POST -d 'username=jon' -d 'password=shhh!' localhost:1323/login
```

*Response*

```js
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjE5NTcxMzZ9.RB3arc4-OyzASAaUhC2W3ReWaXAt_z2Fd3BN4aWTgEY"
}
```

### Request

Request a restricted resource using the token in `Authorization` request header.

```sh
curl localhost:1323/restricted -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjE5NTcxMzZ9.RB3arc4-OyzASAaUhC2W3ReWaXAt_z2Fd3BN4aWTgEY"
```

*Response*

```
Welcome Jon Snow!
```

## Source Code

- [With default Map claims]({{< source "jwt/map-claims" >}})
- [With custom claims]({{< source "jwt/custom-claims" >}})

## Maintainers

- [vishr](https://github.com/vishr)
- [axdg](https://github.com/axdg)
- [matcornic](https://github.com/matcornic)
