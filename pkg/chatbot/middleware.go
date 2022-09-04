package chatbot

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

var subcommands = make(map[string]struct{})

func middlewareMessagePrefixIs(str string, f socketmode.SocketmodeHandlerFunc) socketmode.SocketmodeHandlerFunc {
	if _, exists := subcommands[str]; exists {
		panic(fmt.Sprintf(`subcommand "%s" has already been registered`, str))
	}
	subcommands[str] = struct{}{}
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
		s := strings.Fields(appMentionEvent.Text)
		if len(s) >= 2 && s[1] == str {
			f(evt, c)
		}
	}
}
