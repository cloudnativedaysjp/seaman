package middleware

import (
	"fmt"
	"strings"
	"sync"

	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Command struct {
	Prefixes []string
	URL      string
}

func RegisterCommand(prefixes ...string) *Command {
	return &Command{Prefixes: prefixes}
}

func (m Command) WithURL(url string) *Command {
	m.URL = url
	return &m
}

func (m Command) prefixes() string {
	return strings.Join(m.Prefixes, " ")
}

func (m Command) Handle(next socketmode.SocketmodeHandlerFunc) socketmode.SocketmodeHandlerFunc {
	if Subcommands.Exists(m.prefixes()) {
		panic(fmt.Sprintf(`subcommand "%s" has already been registered`, m.prefixes()))
	}
	Subcommands.Set(m.prefixes(), m.URL)

	// this middleware is intended to be called only incoming slackevents.AppMention.
	// So call panic() if triggered by other incoming.
	panicF := func() {
		panic("cannot use this middleware if it is triggered by other than slackevents.AppMention")
	}
	// return function
	return func(evt *socketmode.Event, c *socketmode.Client) {
		eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			panicF()
		}
		appMentionEvent, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			panicF()
		}
		s := strings.Fields(appMentionEvent.Text)[1:]
		if len(s) >= len(m.Prefixes) && strings.HasPrefix(strings.Join(s, " "), strings.Join(m.Prefixes, " ")) {
			next(evt, c)
		}
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
