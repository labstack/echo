package echotest

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestServeWithHandler(t *testing.T) {
	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, c.QueryParam("key"))
	}
	testConf := ContextConfig{
		QueryValues: url.Values{"key": []string{"value"}},
	}

	resp := testConf.ServeWithHandler(t, handler)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "value", resp.Body.String())
}

func TestServeWithHandler_error(t *testing.T) {
	handler := func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "something went wrong")
	}
	testConf := ContextConfig{
		QueryValues: url.Values{"key": []string{"value"}},
	}

	customErrHandler := echo.DefaultHTTPErrorHandler(true)

	resp := testConf.ServeWithHandler(t, handler, customErrHandler)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, `{"message":"something went wrong"}`+"\n", resp.Body.String())
}

func TestToContext_QueryValues(t *testing.T) {
	testConf := ContextConfig{
		QueryValues: url.Values{"t": []string{"2006-01-02"}},
	}
	c := testConf.ToContext(t)

	v, err := echo.QueryParam[string](c, "t")

	assert.NoError(t, err)
	assert.Equal(t, "2006-01-02", v)
}

func TestToContext_Headers(t *testing.T) {
	testConf := ContextConfig{
		Headers: http.Header{echo.HeaderXRequestID: []string{"ABC"}},
	}
	c := testConf.ToContext(t)

	id := c.Request().Header.Get(echo.HeaderXRequestID)

	assert.Equal(t, "ABC", id)
}

func TestToContext_PathValues(t *testing.T) {
	testConf := ContextConfig{
		PathValues: echo.PathValues{{
			Name:  "key",
			Value: "value",
		}},
	}
	c := testConf.ToContext(t)

	key := c.Param("key")

	assert.Equal(t, "value", key)
}

func TestToContext_RouteInfo(t *testing.T) {
	testConf := ContextConfig{
		RouteInfo: &echo.RouteInfo{
			Name:       "my_route",
			Method:     http.MethodGet,
			Path:       "/:id",
			Parameters: []string{"id"},
		},
	}
	c := testConf.ToContext(t)

	ri := c.RouteInfo()

	assert.Equal(t, echo.RouteInfo{
		Name:       "my_route",
		Method:     http.MethodGet,
		Path:       "/:id",
		Parameters: []string{"id"},
	}, ri)
}

func TestToContext_FormValues(t *testing.T) {
	testConf := ContextConfig{
		FormValues: url.Values{"key": []string{"value"}},
	}
	c := testConf.ToContext(t)

	assert.Equal(t, "value", c.FormValue("key"))
	assert.Equal(t, http.MethodPost, c.Request().Method)
	assert.Equal(t, echo.MIMEApplicationForm, c.Request().Header.Get(echo.HeaderContentType))
}

func TestToContext_MultipartForm(t *testing.T) {
	testConf := ContextConfig{
		MultipartForm: &MultipartForm{
			Fields: map[string]string{
				"key": "value",
			},
			Files: []MultipartFormFile{
				{
					Fieldname: "file",
					Filename:  "test.json",
					Content:   LoadBytes(t, "testdata/test.json"),
				},
			},
		},
	}
	c := testConf.ToContext(t)

	assert.Equal(t, "value", c.FormValue("key"))
	assert.Equal(t, http.MethodPost, c.Request().Method)
	assert.Equal(t, true, strings.HasPrefix(c.Request().Header.Get(echo.HeaderContentType), "multipart/form-data; boundary="))

	fv, err := c.FormFile("file")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test.json", fv.Filename)
	assert.Equal(t, int64(23), fv.Size)
}

func TestToContext_JSONBody(t *testing.T) {
	testConf := ContextConfig{
		JSONBody: LoadBytes(t, "testdata/test.json"),
	}
	c := testConf.ToContext(t)

	payload := struct {
		Field string `json:"field"`
	}{}
	if err := c.Bind(&payload); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "value", payload.Field)
	assert.Equal(t, http.MethodPost, c.Request().Method)
	assert.Equal(t, echo.MIMEApplicationJSON, c.Request().Header.Get(echo.HeaderContentType))
}
