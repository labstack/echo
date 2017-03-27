+++
title = "Testing"
description = "Testing handler and middleware in Echo"
[menu.main]
  name = "Testing"
  parent = "guide"
+++

## Testing Handler

`GET` `/users/:id`

Handler below retrieves user by id from the database. If user is not found it returns
`404` error with a message.

### CreateUser

`POST` `/users`

- Accepts JSON payload
- On success `201 - Created`
- On error `500 - Internal Server Error`

### GetUser

`GET` `/users/:email`

- On success `200 - OK`
- On error `404 - Not Found` if user is not found otherwise `500 - Internal Server Error`

`handler.go`

```go
package handler

import (
	"net/http"

	"github.com/labstack/echo"
)

type (
	User struct {
		Name  string `json:"name" form:"name"`
		Email string `json:"email" form:"email"`
	}
	handler struct {
		db map[string]*User
	}
)

func (h *handler) createUser(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, u)
}

func (h *handler) getUser(c echo.Context) error {
	email := c.Param("email")
	user := h.db[email]
	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return c.JSON(http.StatusOK, user)
}
```

`handler_test.go`

```go
package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

var (
	mockDB = map[string]*User{
		"jon@labstack.com": &User{"Jon Snow", "jon@labstack.com"},
	}
	userJSON = `{"name":"Jon Snow","email":"jon@labstack.com"}`
)

func TestCreateUser(t *testing.T) {
	// Setup
	e := echo.New()
	req, err := http.NewRequest(echo.POST, "/users", strings.NewReader(userJSON))
	if assert.NoError(t, err) {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := &handler{mockDB}

		// Assertions
		if assert.NoError(t, h.createUser(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Equal(t, userJSON, rec.Body.String())
		}
	}
}

func TestGetUser(t *testing.T) {
	// Setup
	e := echo.New()
	req := new(http.Request)
	rec := httptest.NewRecorder()
  c := e.NewContext(req, rec)
	c.SetPath("/users/:email")
	c.SetParamNames("email")
	c.SetParamValues("jon@labstack.com")
	h := &handler{mockDB}

	// Assertions
	if assert.NoError(t, h.getUser(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, userJSON, rec.Body.String())
	}
}
```

### Using Form Payload

```go
f := make(url.Values)
f.Set("name", "Jon Snow")
f.Set("email", "jon@labstack.com")
req, err := http.NewRequest(echo.POST, "/", strings.NewReader(f.Encode()))
```

### Setting Path Params

```go
c.SetParamNames("id", "email")
c.SetParamValues("1", "jon@labstack.com")
```

### Setting Query Params

```go
q := make(url.Values)
q.Set("email", "jon@labstack.com")
req, err := http.NewRequest(echo.POST, "/?"+q.Encode(), nil)
```

## Testing Middleware

Middleware is declared as:

``` go
type MiddlewareFunc func(next HandlerFunc) HandlerFunc
type HandlerFunc func(c Context) error
```

meaning we can make a simple middleware that just checks the user's auth credentials

``` go
func checkForBatman(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, pass, ok := c.Request().BasicAuth()
		if ok && pass == "j0kEr5ux!" && user == "batman" {
			return next(c)
		}
		return echo.ErrUnauthorized
	}
}
```

And to create a test for it you just have to setup the echo and context.

``` go
func TestForBatman(t *testing.T) {
	e := echo.New()

  // create request, recorder and context for the actual query
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)

	// create an endpoint to call and flag to verify it was called
	called := false
	next := func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	}

  // set the basic auth to match our test
	r.SetBasicAuth("batman", "j0kEr5ux!")
	handler := checkForBatman(next)

	// now make the actual call with the context created above
	assert.NoError(t, handler(c))

	// test that we got back what we wanted
	assert.True(t, called)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// reset for a new call with a bad value
	r.SetBasicAuth("joker", "batty")
	called = false
	err := handler(c)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)

	// validate that we didn't call through
	assert.False(t, called)
}
```

For more examples you can look into built-in middleware [test cases](https://github.com/labstack/echo/tree/master/middleware).
