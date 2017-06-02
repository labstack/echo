package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin"
	"github.com/labstack/echo"
)

func testRequest(t *testing.T, ce *casbin.Enforcer, user string, path string, method string, code int) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	req.SetBasicAuth(user, "secret")
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := CasbinAuth(ce)(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	err := h(c)

	if err != nil {
		if errObj, ok := err.(*echo.HTTPError); ok {
			if errObj.Code != code {
				t.Errorf("%s, %s, %s: %d, supposed to be %d", user, path, method, errObj.Code, code)
			}
		} else {
			t.Error(err)
		}
	} else {
		if c.Response().Status != code {
			t.Errorf("%s, %s, %s: %d, supposed to be %d", user, path, method, c.Response().Status, code)
		}
	}
}

func TestCasbinAuth(t *testing.T) {
	ce := casbin.NewEnforcer("casbin_auth_model.conf", "casbin_auth_policy.csv")

	testRequest(t, ce, "alice", "/dataset1/resource1", echo.GET, 200)
	testRequest(t, ce, "alice", "/dataset1/resource1", echo.POST, 200)
	testRequest(t, ce, "alice", "/dataset1/resource2", echo.GET, 200)
	testRequest(t, ce, "alice", "/dataset1/resource2", echo.POST, 403)
}

func TestPathWildcard(t *testing.T) {
	ce := casbin.NewEnforcer("casbin_auth_model.conf", "casbin_auth_policy.csv")

	testRequest(t, ce, "bob", "/dataset2/resource1", "GET", 200)
	testRequest(t, ce, "bob", "/dataset2/resource1", "POST", 200)
	testRequest(t, ce, "bob", "/dataset2/resource1", "DELETE", 200)
	testRequest(t, ce, "bob", "/dataset2/resource2", "GET", 200)
	testRequest(t, ce, "bob", "/dataset2/resource2", "POST", 403)
	testRequest(t, ce, "bob", "/dataset2/resource2", "DELETE", 403)

	testRequest(t, ce, "bob", "/dataset2/folder1/item1", "GET", 403)
	testRequest(t, ce, "bob", "/dataset2/folder1/item1", "POST", 200)
	testRequest(t, ce, "bob", "/dataset2/folder1/item1", "DELETE", 403)
	testRequest(t, ce, "bob", "/dataset2/folder1/item2", "GET", 403)
	testRequest(t, ce, "bob", "/dataset2/folder1/item2", "POST", 200)
	testRequest(t, ce, "bob", "/dataset2/folder1/item2", "DELETE", 403)
}

func TestRBAC(t *testing.T) {
	ce := casbin.NewEnforcer("casbin_auth_model.conf", "casbin_auth_policy.csv")

	// cathy can access all /dataset1/* resources via all methods because it has the dataset1_admin role.
	testRequest(t, ce, "cathy", "/dataset1/item", "GET", 200)
	testRequest(t, ce, "cathy", "/dataset1/item", "POST", 200)
	testRequest(t, ce, "cathy", "/dataset1/item", "DELETE", 200)
	testRequest(t, ce, "cathy", "/dataset2/item", "GET", 403)
	testRequest(t, ce, "cathy", "/dataset2/item", "POST", 403)
	testRequest(t, ce, "cathy", "/dataset2/item", "DELETE", 403)

	// delete all roles on user cathy, so cathy cannot access any resources now.
	ce.DeleteRolesForUser("cathy")

	testRequest(t, ce, "cathy", "/dataset1/item", "GET", 403)
	testRequest(t, ce, "cathy", "/dataset1/item", "POST", 403)
	testRequest(t, ce, "cathy", "/dataset1/item", "DELETE", 403)
	testRequest(t, ce, "cathy", "/dataset2/item", "GET", 403)
	testRequest(t, ce, "cathy", "/dataset2/item", "POST", 403)
	testRequest(t, ce, "cathy", "/dataset2/item", "DELETE", 403)
}
