package lacks

import (
	"context"
	"fmt"

	"github.com/cloudnativedaysjp/seaman/pkg/log"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type funcAppMentionEventForHelp func(context.Context, *slackevents.AppMentionEvent, *socketmode.Client, map[string]string)

func (h *handler) HandleHelp(callback funcAppMentionEventForHelp) {
	ctx := context.Background()

	h.socketmodeHandler.HandleEvents(slackevents.AppMention, func(evt *socketmode.Event, client *socketmode.Client) {
		client.Ack(*evt.Request)

		ev, err := getAppMentionEvent(evt)
		if err != nil {
			h.log.Error(fmt.Sprintf("failed to get InteractionCallback: %v", err))
			return
		}
		ctx = log.IntoContext(ctx, h.log.With("messageTs", ev.TimeStamp))

		callback(ctx, ev, client, h.commands)
	})
}
