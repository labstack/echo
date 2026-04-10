// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v5"
)

// MCPConfig defines the config for MCP middleware. The middleware mounts a
// Model Context Protocol (https://modelcontextprotocol.io) JSON-RPC 2.0 endpoint
// at MCPConfig.Path and exposes every route registered on the Echo instance as
// an MCP "tool" that an AI client (Claude Desktop, Cursor, VS Code, ...) can
// discover via tools/list and invoke via tools/call.
type MCPConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Name is the server name advertised in the initialize handshake.
	// Default value "echo-mcp".
	Name string

	// Version is the server version advertised in the initialize handshake.
	// Default value "0.0.0".
	Version string

	// Path is the URL path the MCP JSON-RPC endpoint is mounted on. Only POST
	// requests to this exact path are handled; every other request is passed
	// through to the next handler unchanged.
	// Default value "/mcp".
	Path string
}

// MCP returns a middleware that exposes registered Echo routes as MCP tools at
// the path configured in MCPConfig.Path.
func MCP(config MCPConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts MCPConfig to middleware or returns an error for invalid configuration.
func (config MCPConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.Path == "" {
		config.Path = "/mcp"
	}
	if config.Name == "" {
		config.Name = "echo-mcp"
	}
	if config.Version == "" {
		config.Version = "0.0.0"
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			if c.Request().URL.Path != config.Path {
				return next(c)
			}
			if c.Request().Method != http.MethodPost {
				return next(c)
			}
			return handleMCP(c, config)
		}
	}, nil
}

// --- JSON-RPC envelope types ------------------------------------------------

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// --- request handling -------------------------------------------------------

func handleMCP(c *echo.Context, cfg MCPConfig) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusOK, rpcResponse{
			JSONRPC: "2.0",
			Error:   &rpcError{Code: -32700, Message: "parse error: " + err.Error()},
		})
	}

	var req rpcRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return c.JSON(http.StatusOK, rpcResponse{
			JSONRPC: "2.0",
			Error:   &rpcError{Code: -32700, Message: "parse error: " + err.Error()},
		})
	}

	switch req.Method {
	case "initialize":
		return c.JSON(http.StatusOK, rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    cfg.Name,
					"version": cfg.Version,
				},
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
			},
		})

	case "notifications/initialized":
		// Notifications carry no ID and expect no response.
		return c.NoContent(http.StatusNoContent)

	case "tools/list":
		tools, _ := buildTools(c.Echo().Router().Routes(), cfg.Path)
		return c.JSON(http.StatusOK, rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{"tools": tools},
		})

	case "tools/call":
		result, err := callTool(c, cfg, req.Params)
		if err != nil {
			return c.JSON(http.StatusOK, rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &rpcError{Code: -32603, Message: err.Error()},
			})
		}
		return c.JSON(http.StatusOK, rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		})

	default:
		return c.JSON(http.StatusOK, rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32601, Message: "method not found: " + req.Method},
		})
	}
}

// --- tool building ----------------------------------------------------------

// buildTools turns Echo's registered routes into MCP tool descriptors and
// returns a parallel name->RouteInfo map used by callTool to look up the route
// to dispatch to.
func buildTools(routes echo.Routes, mcpPath string) ([]map[string]any, map[string]echo.RouteInfo) {
	tools := make([]map[string]any, 0, len(routes))
	index := make(map[string]echo.RouteInfo, len(routes))
	used := make(map[string]int, len(routes))

	for _, ri := range routes {
		if ri.Path == mcpPath {
			continue // never expose the MCP endpoint itself
		}

		name := toolName(ri)
		if used[name] > 0 {
			name = fmt.Sprintf("%s_%d", name, used[name]+1)
		}
		used[toolName(ri)]++
		index[name] = ri

		properties := map[string]any{}
		required := make([]string, 0, len(ri.Parameters))
		for _, p := range ri.Parameters {
			properties[p] = map[string]any{
				"type":        "string",
				"description": "Path parameter :" + p,
			}
			required = append(required, p)
		}
		properties["query"] = map[string]any{
			"type":                 "object",
			"description":          "Optional query string parameters as a flat key/value object.",
			"additionalProperties": true,
		}
		if methodHasBody(ri.Method) {
			properties["body"] = map[string]any{
				"type":                 "object",
				"description":          "JSON request body.",
				"additionalProperties": true,
			}
		}

		schema := map[string]any{
			"type":       "object",
			"properties": properties,
		}
		if len(required) > 0 {
			schema["required"] = required
		}

		tools = append(tools, map[string]any{
			"name":        name,
			"description": fmt.Sprintf("%s %s", ri.Method, ri.Path),
			"inputSchema": schema,
		})
	}

	return tools, index
}

func toolName(ri echo.RouteInfo) string {
	if ri.Name != "" && ri.Name != ri.Method+":"+ri.Path {
		return sanitize(ri.Name)
	}
	slug := strings.NewReplacer("/", "_", ":", "", "*", "wild").Replace(ri.Path)
	slug = strings.Trim(slug, "_")
	if slug == "" {
		slug = "root"
	}
	return ri.Method + "_" + slug
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func methodHasBody(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

// --- tool invocation --------------------------------------------------------

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func callTool(c *echo.Context, cfg MCPConfig, raw json.RawMessage) (any, error) {
	var p toolCallParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("invalid tools/call params: %w", err)
	}
	if p.Name == "" {
		return nil, errors.New("tools/call: missing tool name")
	}

	_, index := buildTools(c.Echo().Router().Routes(), cfg.Path)
	ri, ok := index[p.Name]
	if !ok {
		return nil, fmt.Errorf("tools/call: unknown tool %q", p.Name)
	}

	// Substitute path parameters in order using RouteInfo.Reverse.
	pathValues := make([]any, 0, len(ri.Parameters))
	for _, name := range ri.Parameters {
		v, ok := p.Arguments[name]
		if !ok {
			return nil, fmt.Errorf("tools/call: missing required argument %q", name)
		}
		pathValues = append(pathValues, v)
	}
	target := ri.Reverse(pathValues...)

	// Build query string from arguments["query"] if present.
	if q, ok := p.Arguments["query"].(map[string]any); ok && len(q) > 0 {
		values := url.Values{}
		for k, v := range q {
			values.Set(k, fmt.Sprintf("%v", v))
		}
		if strings.Contains(target, "?") {
			target += "&" + values.Encode()
		} else {
			target += "?" + values.Encode()
		}
	}

	// Build body from arguments["body"] if present.
	var bodyReader io.Reader
	if b, ok := p.Arguments["body"]; ok && b != nil {
		raw, err := json.Marshal(b)
		if err != nil {
			return nil, fmt.Errorf("tools/call: cannot marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	innerReq, err := http.NewRequestWithContext(c.Request().Context(), ri.Method, target, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("tools/call: cannot build request: %w", err)
	}
	// http.NewRequest leaves RequestURI empty (it is only set on server-side
	// requests), but Echo's request logger and other middlewares read it.
	// Populate it so the synthesized request looks like a real one downstream.
	innerReq.RequestURI = target
	if bodyReader != nil {
		innerReq.Header.Set("Content-Type", "application/json")
	}
	innerReq.Header.Set("Accept", "application/json")

	rw := &bufferRW{header: http.Header{}}
	c.Echo().ServeHTTP(rw, innerReq)

	text := rw.body.String()
	if text == "" {
		text = fmt.Sprintf("(status %d, empty body)", rw.statusOrDefault())
	}

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
		"isError": rw.statusOrDefault() >= 400,
	}, nil
}

// --- in-memory http.ResponseWriter -----------------------------------------

type bufferRW struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (w *bufferRW) Header() http.Header { return w.header }

func (w *bufferRW) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
}

func (w *bufferRW) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(b)
}

func (w *bufferRW) statusOrDefault() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}
