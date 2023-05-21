package lacks

import (
	"context"
	"fmt"

	"github.com/cloudnativedaysjp/seaman/pkg/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type funcInteractionCallback func(context.Context, slack.InteractionCallback, *socketmode.Client)

func (h *handler) HandleInteractionBlockAction(actionID string, callback funcInteractionCallback) {
	ctx := context.Background()
	h.socketmodeHandler.HandleInteractionBlockAction(actionID, func(evt *socketmode.Event, client *socketmode.Client) {
		client.Ack(*evt.Request)

		interaction, err := getInteractionCallback(evt)
		if err != nil {
			h.log.Error(fmt.Sprintf("failed to get InteractionCallback: %v", err))
			return
		}
		ctx = log.IntoContext(ctx, h.log.With("messageTs", interaction.Container.MessageTs))

		callback(ctx, interaction, client)
	})
}
