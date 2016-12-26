+++
title = "Request"
description = "Handling HTTP request in Echo"
[menu.main]
  name = "Request"
  parent = "guide"
  weight = 6
+++

## Bind Data

To bind request body into a Go type use `Context#Bind(i interface{})`.
The default binder supports decoding application/json, application/xml and
application/x-www-form-urlencoded data based on the Content-Type header.

Example below binds the request payload into `User` struct based on tags:

```go
// User
User struct {
  Name  string `json:"name" form:"name" query:"name"`
  Email string `json:"email" form:"email" query:"email"`
}
```

```go
// Handler
func(c echo.Context) (err error) {
  u := new(User)
  if err = c.Bind(u); err != nil {
    return
  }
  return c.JSON(http.StatusOK, u)
}
```

### JSON Data

```sh
curl \
  -X POST \
  http://localhost:1323/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Joe","email":"joe@labstack"}'
```

### Form Data

```sh
curl \
  -X POST \
  http://localhost:1323/users \
  -d 'name=Joe' \
  -d 'email=joe@labstack.com'
```

### Query Parameters

```sh
curl \
  -X GET \
  http://localhost:1323/users\?name\=Joe\&email\=joe@labstack.com
```

## Custom Binder

Custom binder can be registered using `Echo#Binder`.

*Example*

```go
type CustomBinder struct {}

func (cb *CustomBinder) Bind(i interface{}, c echo.Context) (err error) {
	// You may use default binder
	db := new(echo.DefaultBinder)
	if err = db.Bind(i, c); err != echo.ErrUnsupportedMediaType {
		return
	}

	// Define your custom implementation

	return
}
```

## Retrieve Data

### Form Data

Form data can be retrieved by name using `Context#FormValue(name string)`.

*Example*

```go
// Handler
func(c echo.Context) error {
	name := c.FormValue("name")
	return c.String(http.StatusOK, name)
}
```

```sh
curl \
  -X POST \
  http://localhost:1323 \
  -d 'name=Joe'
```

To bind a custom data type, you can implement `Echo#BindUnmarshaler` interface.

*Example*

```go
type Timestamp time.Time

func (t *Timestamp) UnmarshalParam(src string) error {
	ts, err := time.Parse(time.RFC3339, src)
	*t = Timestamp(ts)
	return err
}
```

### Query Parameters

Query parameters can be retrieved by name using `Context#QueryParam(name string)`.

*Example*

```go
// Handler
func(c echo.Context) error {
	name := c.QueryParam("name")
	return c.String(http.StatusOK, name)
})
```

```sh
curl \
  -X GET \
  http://localhost:1323\?name\=Joe
```

Similar to form data, custom data type can be bind using `Context#QueryParam(name string)`.

### Path Parameters

Registered path parameters can be retrieved by name using `Context#Param(name string) string`.

*Example*

```go
e.GET("/users/:name", func(c echo.Context) error {
	name := c.Param("name")
	return c.String(http.StatusOK, name)
})
```

```sh
$ curl http://localhost:1323/users/Joe
```

## Validate Data

Echo doesn't have a built-in data validation capabilities, however, you can register
a custom validator using `Echo#Validator` and leverage third-party [libraries](https://github.com/avelino/awesome-go#validation).

Example below uses https://github.com/go-playground/validator framework for validation:

```go
type (
	User struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	CustomValidator struct {
		validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.POST("/users", func(c echo.Context) (err error) {
		u := new(User)
		if err = c.Bind(u); err != nil {
			return
		}
		if err = c.Validate(u); err != nil {
			return
		}
		return c.JSON(http.StatusOK, u)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
```

```sh
curl \
  -X POST \
  http://localhost:1323/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Joe","email":"joe@invalid-domain"}'
{"message":"Key: 'User.Email' Error:Field validation for 'Email' failed on the 'email' tag"}
```
