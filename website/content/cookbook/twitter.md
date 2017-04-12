+++
title = "Twitter Like API"
description = "Twitter Like API example for Echo"
[menu.main]
  name = "Twitter"
  parent = "cookbook"
+++

This example shows how to create a Twitter like REST API using MongoDB (Database),
JWT (API security) and JSON (Data exchange).

## Models

`user.go`

{{< embed "twitter/model/user.go" >}}

`post.go`

{{< embed "twitter/model/post.go" >}}

## Handlers

`handler.go`

{{< embed "twitter/handler/handler.go" >}}

`user.go`

{{< embed "twitter/handler/user.go" >}}

`post.go`

{{< embed "twitter/handler/post.go" >}}

## Server

`server.go`

{{< embed "twitter/server.go" >}}

## API

### Signup

User signup

- Retrieve user credentials from the body and validate against database.
- For invalid email or password, send `400 - Bad Request` response.
- For valid email and password, save user in database and send `201 - Created` response.

#### Request

```sh
curl \
  -X POST \
  http://localhost:1323/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"jon@labstack.com","password":"shhh!"}'
```

#### Response

`201 - Created`

```js
{
  "id": "58465b4ea6fe886d3215c6df",
  "email": "jon@labstack.com",
  "password": "shhh!"
}
```

### Login

User login

- Retrieve user credentials from the body and validate against database.
- For invalid credentials, send `401 - Unauthorized` response.
- For valid credentials, send `200 - OK` response:
  - Generate JWT for the user and send it as response.
  - Each subsequent request must include JWT in the `Authorization` header.

Method: `POST`<br>
Path: `/login`

#### Request

```sh
curl \
  -X POST \
  http://localhost:1323/login \
  -H "Content-Type: application/json" \
  -d '{"email":"jon@labstack.com","password":"shhh!"}'
```

#### Response

`200 - OK`

```js
{
  "id": "58465b4ea6fe886d3215c6df",
  "email": "jon@labstack.com",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODEyNjUxMjgsImlkIjoiNTg0NjViNGVhNmZlODg2ZDMyMTVjNmRmIn0.1IsGGxko1qMCsKkJDQ1NfmrZ945XVC9uZpcvDnKwpL0"
}
```

Client should store the token, for browsers, you may use local storage.

### Follow

Follow a user

- For invalid token, send `400 - Bad Request` response.
- For valid token:
  - If user is not found, send `404 - Not Found` response.
  - Add a follower to the specified user in the path parameter and send `200 - OK` response.

Method: `POST` <br>
Path: `/follow/:id`

#### Request

```sh
curl \
  -X POST \
  http://localhost:1323/follow/58465b4ea6fe886d3215c6df \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODEyNjUxMjgsImlkIjoiNTg0NjViNGVhNmZlODg2ZDMyMTVjNmRmIn0.1IsGGxko1qMCsKkJDQ1NfmrZ945XVC9uZpcvDnKwpL0"
```

#### Response

`200 - OK`

### Post

Post a message to specified user

- For invalid request payload, send `400 - Bad Request` response.
- If user is not found, send `404 - Not Found` response.
- Otherwise save post in the database and return it via `201 - Created` response.

Method: `POST` <br>
Path: `/posts`

#### Request

```sh
curl \
  -X POST \
  http://localhost:1323/posts \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODEyNjUxMjgsImlkIjoiNTg0NjViNGVhNmZlODg2ZDMyMTVjNmRmIn0.1IsGGxko1qMCsKkJDQ1NfmrZ945XVC9uZpcvDnKwpL0" \
  -H "Content-Type: application/json" \
  -d '{"to":"58465b4ea6fe886d3215c6df","message":"hello"}'
```

#### Response

`201 - Created`

```js
{
  "id": "584661b9a6fe8871a3804cba",
  "to": "58465b4ea6fe886d3215c6df",
  "from": "58465b4ea6fe886d3215c6df",
  "message": "hello"
}
```

### Feed

List most recent messages based on optional `page` and `limit` query parameters

Method: `GET` <br>
Path: `/feed?page=1&limit=5`

#### Request

```sh
curl \
  -X GET \
  http://localhost:1323/feed \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODEyNjUxMjgsImlkIjoiNTg0NjViNGVhNmZlODg2ZDMyMTVjNmRmIn0.1IsGGxko1qMCsKkJDQ1NfmrZ945XVC9uZpcvDnKwpL0"
```

#### Response

`200 - OK`

```js
[
  {
    "id": "584661b9a6fe8871a3804cba",
    "to": "58465b4ea6fe886d3215c6df",
    "from": "58465b4ea6fe886d3215c6df",
    "message": "hello"
  }
]
```

## [Source Code]({{< source "twitter" >}})

## Maintainers

- [vishr](https://github.com/vishr)
