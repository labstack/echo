// +build !go1.7

package context

import "golang.org/x/net/context"

func Background() Context {
	return context.Background()
}

func WithValue(parent Context, key, val interface{}) Context {
	return context.WithValue(parent, key, val)
}
