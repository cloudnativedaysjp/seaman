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

type funcAppMentionEvent func(context.Context, *slackevents.AppMentionEvent, *socketmode.Client) error

func (h *router) HandleMentionedMessage(c string, callback funcAppMentionEvent) OptBuilderMentionedMessage {
	ctx := context.Background()
	h.commands = append(h.commands, command{prefixes: strings.Fields(c)})

	h.socketmodeHandler.HandleEvents(slackevents.AppMention, func(evt *socketmode.Event, client *socketmode.Client) {
		ev, err := utils.GetAppMentionEvent(evt)
		inputCmds := strings.Join(strings.Fields(ev.Text)[1:], " ")
		if !strings.HasPrefix(inputCmds, c) {
			return
		}

		client.Ack(*evt.Request)

		if err != nil {
			h.log.Error(fmt.Sprintf("failed to get InteractionCallback: %v", err))
			return
		}
		ctx = log.IntoContext(ctx, h.log.
			With("messageTs", ev.TimeStamp).
			With("input", inputCmds),
		)

		if err := callback(ctx, ev, client); err != nil {
			h.log.Error(err.Error(), log.KeyDetail, err)
		}
	})
	return OptBuilderMentionedMessage{h, c}
}

type OptBuilderMentionedMessage struct {
	owner   *router
	command string
}

func (builder OptBuilderMentionedMessage) WithURL(url string) {
	for i, c := range builder.owner.commands {
		if c.prefix() == builder.command {
			builder.owner.commands[i] = command{prefixes: c.prefixes, url: url}
		}
	}
}
