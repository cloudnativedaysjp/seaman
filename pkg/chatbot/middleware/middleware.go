package middleware

import (
	"github.com/slack-go/slack/socketmode"
)

type Middleware interface {
	Handle(next socketmode.SocketmodeHandlerFunc) socketmode.SocketmodeHandlerFunc
}

// MiddlewareSet gather Middlewares and return SocketmodeHandlerFunc
func MiddlewareSet(h socketmode.SocketmodeHandlerFunc, middlewares ...Middleware) socketmode.SocketmodeHandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i].Handle(h)
	}
	return h
}
