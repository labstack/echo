// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import "strings"

// routePathPart is one parsed piece of a route pattern. A backslash before a
// colon makes the colon static, including when it follows a parameter name.
type routePathPart struct {
	kind  kind
	value string
}

func parseRoutePath(path string) []routePathPart {
	var parts []routePathPart
	var literal strings.Builder
	flushLiteral := func() {
		if literal.Len() > 0 {
			parts = append(parts, routePathPart{kind: staticKind, value: literal.String()})
			literal.Reset()
		}
	}

	for i := 0; i < len(path); {
		switch {
		case path[i] == '\\' && i+1 < len(path) && path[i+1] == ':':
			literal.WriteByte(':')
			i += 2
		case path[i] == ':':
			flushLiteral()
			start := i + 1
			i = start
			for i < len(path) && path[i] != '/' {
				if path[i] == '\\' && i+1 < len(path) && path[i+1] == ':' {
					break
				}
				i++
			}
			parts = append(parts, routePathPart{kind: paramKind, value: path[start:i]})
		case path[i] == '*':
			flushLiteral()
			start := i
			for i < len(path) && path[i] != '/' {
				i++
			}
			parts = append(parts, routePathPart{kind: anyKind, value: path[start:i]})
		default:
			literal.WriteByte(path[i])
			i++
		}
	}
	flushLiteral()
	return parts
}

func routeTreePath(parts []routePathPart) (string, []int) {
	var path strings.Builder
	var paramMarkers []int
	for _, part := range parts {
		switch part.kind {
		case staticKind:
			path.WriteString(part.value)
		case paramKind:
			paramMarkers = append(paramMarkers, path.Len())
			path.WriteByte(paramLabel)
		case anyKind:
			path.WriteByte(anyLabel)
			return path.String(), paramMarkers
		}
	}
	return path.String(), paramMarkers
}
