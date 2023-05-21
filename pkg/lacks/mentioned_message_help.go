package lacks

import (
	"context"
	"fmt"
	"strings"

	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/cloudnativedaysjp/seaman/pkg/log"
	"github.com/cloudnativedaysjp/seaman/pkg/utils"
)

type funcAppMentionEventForHelp func(context.Context, *slackevents.AppMentionEvent, *socketmode.Client, map[string]string) error

func (h *router) HandleHelp(callback funcAppMentionEventForHelp) {
	ctx := context.Background()

	h.socketmodeHandler.HandleEvents(slackevents.AppMention, func(evt *socketmode.Event, client *socketmode.Client) {
		ev, err := utils.GetAppMentionEvent(evt)
		inputCmds := strings.Join(strings.Fields(ev.Text)[1:], " ")
		if !strings.HasPrefix(inputCmds, "help") {
			return
		}

		client.Ack(*evt.Request)

		if err != nil {
			h.log.Error(fmt.Sprintf("failed to get InteractionCallback: %v", err))
			return
		}
		ctx = log.IntoContext(ctx, h.log.
			With("messageTs", ev.TimeStamp).
			With("commands", "help"),
		)

		commands := make(map[string]string)
		for _, cmd := range h.commands {
			commands[cmd.prefix()] = cmd.url
		}

		if err := callback(ctx, ev, client, commands); err != nil {
			h.log.Error(err.Error(), log.KeyDetail, err)
		}
	})
}
