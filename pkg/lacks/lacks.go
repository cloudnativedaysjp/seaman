package lacks

import (
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/exp/slog"
)

type handler struct {
	childs            []*handler
	commands          map[string]string
	log               *slog.Logger
	socketmodeHandler *socketmode.SocketmodeHandler
}

func New(logger *slog.Logger, client *socketmode.Client) *handler {
	return &handler{
		[]*handler{},
		make(map[string]string),
		logger,
		socketmode.NewSocketmodeHandler(client),
	}
}

func (h *handler) RunEventLoop() error {
	return h.socketmodeHandler.RunEventLoop()
}
