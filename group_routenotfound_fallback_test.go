package echo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupRouteNotFoundFallsBackToRootHandler(t *testing.T) {
	e := New()
	e.HideBanner = true
	e.HidePort = true

	e.RouteNotFound("/*", func(c Context) error {
		return c.NoContent(http.StatusNotFound)
	})

	v0 := e.Group("/v0", func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	})
	v0.POST("/resource", func(c Context) error {
		return c.NoContent(http.StatusOK)
	})

	v1 := e.Group("/v1")
	v1.POST("/resource", func(c Context) error {
		return c.NoContent(http.StatusOK)
	})

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	for _, path := range []string{"/foo", "/v0/foo", "/v1/foo"} {
		t.Run(path, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, srv.URL+path, nil)
			require.NoError(t, err)

			resp, err := srv.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, http.StatusNotFound, resp.StatusCode)
			require.Empty(t, b)
			require.Equal(t, "0", resp.Header.Get("Content-Length"))
		})
	}
}
