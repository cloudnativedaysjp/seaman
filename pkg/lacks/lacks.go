package lacks

import (
	"strings"

	"github.com/slack-go/slack/socketmode"
	"golang.org/x/exp/slog"
)

type router struct {
	commands          []command
	log               *slog.Logger
	socketmodeHandler *socketmode.SocketmodeHandler
}

func NewRouter(logger *slog.Logger, client *socketmode.Client) *router {
	return &router{
		[]command{},
		logger,
		socketmode.NewSocketmodeHandler(client),
	}
}

func (r *router) RunEventLoop() error {
	return r.socketmodeHandler.RunEventLoop()
}

type command struct {
	prefixes []string
	url      string
}

func (c command) prefix() string {
	return strings.Join(c.prefixes, " ")
}
