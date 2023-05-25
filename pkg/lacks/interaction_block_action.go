package lacks

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	"github.com/cloudnativedaysjp/seaman/pkg/log"
	"github.com/cloudnativedaysjp/seaman/pkg/utils"
)

type funcInteractionCallback func(context.Context, slack.InteractionCallback, *socketmode.Client) error

func (h *router) HandleInteractionBlockAction(actionID string, callback funcInteractionCallback) {
	ctx := context.Background()
	h.socketmodeHandler.HandleInteractionBlockAction(actionID, func(evt *socketmode.Event, client *socketmode.Client) {
		client.Ack(*evt.Request)

		interaction, err := utils.GetInteractionCallback(evt)
		if err != nil {
			h.log.Error(fmt.Sprintf("failed to get InteractionCallback: %v", err))
			return
		}
		ctx = log.IntoContext(ctx, h.log.
			With("messageTs", interaction.Container.MessageTs).
			With("callbackValue", utils.GetCallbackValueOnButton(interaction)),
		)

		if err := callback(ctx, interaction, client); err != nil {
			h.log.Error(err.Error(), log.KeyDetail, err)
		}
	})
}
