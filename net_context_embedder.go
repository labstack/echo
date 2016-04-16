package echo

import (
	"time"

	netContext "golang.org/x/net/context"
)

// NetContextEmbedder elevates enables embedding of net.Context operations.
type NetContextEmbedder struct {
	Ctx netContext.Context
}

// SetNetContext sets the context
func (n *NetContextEmbedder) SetNetContext(ctx netContext.Context) {
	n.Ctx = ctx
}

// NetContext retrieves the context
func (n *NetContextEmbedder) NetContext() netContext.Context {
	return n.Ctx
}

// Deadline sets a deadline on the context
func (n *NetContextEmbedder) Deadline() (deadline time.Time, ok bool) {
	return n.Ctx.Deadline()
}

// Done gets the done channel for this request
func (n *NetContextEmbedder) Done() <-chan struct{} {
	return n.Ctx.Done()
}

// Err returns the error in the context
func (n *NetContextEmbedder) Err() error {
	return n.Ctx.Err()
}

// Value retrieves a value from the context
func (n *NetContextEmbedder) Value(key interface{}) interface{} {
	return n.Ctx.Value(key)
}
