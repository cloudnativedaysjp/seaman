package lacks

import (
	"context"
	"fmt"

	"github.com/cloudnativedaysjp/seaman/pkg/log"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type funcAppMentionEvent func(context.Context, *slackevents.AppMentionEvent, *socketmode.Client)

func (h *handler) HandleMentionedMessage(command string, callback funcAppMentionEvent) OptBuilderMentionedMessage {
	ctx := context.Background()
	h.commands[command] = ""

	h.socketmodeHandler.HandleEvents(slackevents.AppMention, func(evt *socketmode.Event, client *socketmode.Client) {
		client.Ack(*evt.Request)

		ev, err := getAppMentionEvent(evt)
		if err != nil {
			h.log.Error(fmt.Sprintf("failed to get InteractionCallback: %v", err))
			return
		}
		ctx = log.IntoContext(ctx, h.log.With("messageTs", ev.TimeStamp))

		callback(ctx, ev, client)
	})
	return OptBuilderMentionedMessage{h, command}
}

type OptBuilderMentionedMessage struct {
	owner   *handler
	command string
}

func (builder OptBuilderMentionedMessage) WithURL(url string) {
	builder.owner.commands[builder.command] = url
}
