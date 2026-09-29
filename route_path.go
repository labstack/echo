// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import "strings"

// routePathPart is one parsed piece of a route pattern. A backslash before a
// colon makes the colon static. After a parameter name it starts an inline verb
// (`/:name\:cancel`) when the rest of that path segment is static.
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
		if isEscapedColon(path, i) {
			emit(routePathPart{kind: staticKind, value: ":"})
			i += 2
		} else if path[i] == ':' {
			start := i + 1
			i = start
			plainName := true // an escaped colon only starts an inline verb after a name without ':' or '*'
			for i < len(path) && path[i] != '/' {
				if isEscapedColon(path, i) {
					if plainName && isInlineVerb(path[i+2:]) {
						break
					}
					// not an inline verb: the rest of the segment is the param name, as before inline verbs
					for i < len(path) && path[i] != '/' {
						i++
					}
					break
				}
				if path[i] == ':' || path[i] == '*' {
					plainName = false
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
			for i < len(path) && path[i] != ':' && path[i] != '*' && !isEscapedColon(path, i) {
				i++
			}
			emit(routePathPart{kind: staticKind, value: path[start:i]})
		}
	}
}

func isEscapedColon(path string, i int) bool {
	return path[i] == '\\' && i+1 < len(path) && path[i+1] == ':'
}

// isInlineVerb reports whether the route text after an escaped colon stays
// static up to the end of its path segment. Only then can the router find where
// the parameter value ends by trying the colons in the requested segment. Other
// escaped colons keep the older meaning and remain part of the parameter name.
func isInlineVerb(rest string) bool {
	for i := 0; i < len(rest) && rest[i] != '/'; i++ {
		switch rest[i] {
		case '*':
			return false
		case ':':
			if i == 0 || rest[i-1] != '\\' {
				return false
			}
		}
	}
	return true
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
			path.WriteByte(':')
		case anyKind:
			path.WriteByte(anyLabel)
			return path.String(), paramMarkers
		}
	}
	return path.String(), paramMarkers
}
