package main

import (
    "net/http"

    "github.com/labstack/echo"
)

type Hosts map[string]http.Handler

// Implement a ServeHTTP method for Hosts.
func (h Hosts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Check for host x in the Hosts map.
    if handler := h[r.Host]; handler != nil {
        handler.ServeHTTP(w, r)
    } else {
        // No handler is registered so redirect.
        // Note that rediret may not be appropriate here.
        // This is just for illustration, else branch could just as easily return
        // an error for bad port, notFound for bad subdomain etc.
        http.Redirect(w, r, "http://localhost:8080", http.StatusTemporaryRedirect)
    }
}

func main() {

    // Some subdomains.
    site := echo.New()
    api := echo.New()

    // Host maps.
    hosts := make(Hosts)
    hosts["localhost:8080"] = site
    hosts["api.localhost:8080"] = api

    // Handler for subdomain a.
    site.Get("/", func(c *echo.Context) error {
        c.String(http.StatusOK, "Website service...")
        return nil
    })

    // Handler for subdomain b.
    api.Get("/", func(c *echo.Context) error {
        c.String(http.StatusOK, "Welcome to the api...")
        return nil
    })

    http.ListenAndServe(":8080", hosts)
}
