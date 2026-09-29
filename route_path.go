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
	walkRoutePath(path, func(part routePathPart) { parts = append(parts, part) })
	return parts
}

// walkRoutePath is the common syntax scanner. Reverse uses it directly to
// avoid allocating a parts slice for each URL it builds.
func walkRoutePath(path string, emit func(routePathPart)) {
	for i := 0; i < len(path); {
		if path[i] == '\\' && i+1 < len(path) && path[i+1] == ':' {
			emit(routePathPart{kind: staticKind, value: ":"})
			i += 2
		} else if path[i] == ':' {
			start := i + 1
			i = start
			for i < len(path) && path[i] != '/' {
				if path[i] == '\\' && i+1 < len(path) && path[i+1] == ':' {
					break
				}
				i++
			}
			emit(routePathPart{kind: paramKind, value: path[start:i]})
		} else if path[i] == '*' {
			start := i
			for i < len(path) && path[i] != '/' {
				i++
			}
			emit(routePathPart{kind: anyKind, value: path[start:i]})
		} else {
			start := i
			for i < len(path) && path[i] != ':' && path[i] != '*' {
				if path[i] == '\\' && i+1 < len(path) && path[i+1] == ':' {
					break
				}
				i++
			}
			emit(routePathPart{kind: staticKind, value: path[start:i]})
		}
	}
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
