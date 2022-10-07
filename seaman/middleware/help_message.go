package middleware

import (
	"fmt"
	"sync"

	"github.com/slack-go/slack/socketmode"
)

type HelpMessage struct {
	Prefix string
	URL    string
}

func (m HelpMessage) Handle(next socketmode.SocketmodeHandlerFunc) socketmode.SocketmodeHandlerFunc {
	if Subcommands.Exists(m.Prefix) {
		panic(fmt.Sprintf(`subcommand "%s" has already been registered`, m.Prefix))
	}
	Subcommands.Set(m.Prefix, m.URL)

	// return function
	return func(evt *socketmode.Event, c *socketmode.Client) {
		next(evt, c)
	}
}

// subcommands is recorded by HelpMessage
type subcommands map[string]string

var (
	Subcommands      subcommands = map[string]string{}
	subcommandsMutex sync.RWMutex
)

func (s subcommands) Set(commandName, url string) {
	subcommandsMutex.Lock()
	defer subcommandsMutex.Unlock()
	if _, exist := Subcommands[commandName]; !exist {
		Subcommands[commandName] = url
	}
}

func (s subcommands) List() map[string]string {
	subcommandsMutex.RLock()
	defer subcommandsMutex.RUnlock()
	return Subcommands
}

func (s subcommands) Exists(commandName string) bool {
	subcommandsMutex.RLock()
	defer subcommandsMutex.RUnlock()
	_, exist := Subcommands[commandName]
	return exist
}
