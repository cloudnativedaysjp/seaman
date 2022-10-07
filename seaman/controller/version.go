package controller

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/slack-go/slack/socketmode"

	infra_slack "github.com/cloudnativedaysjp/seaman/seaman/infra/slack"
	"github.com/cloudnativedaysjp/seaman/seaman/utils"
	"github.com/cloudnativedaysjp/seaman/seaman/view"
)

type VersionController struct {
	slackFactory infra_slack.SlackClientFactory
	log          logr.Logger
}

func NewVersionController(
	logger logr.Logger,
	slackFactory infra_slack.SlackClientFactory,
) *VersionController {
	return &VersionController{slackFactory, logger}
}

func (c *VersionController) ShowVersion(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)

	ev, err := getAppMentionEvent(evt)
	if err != nil {
		c.log.Error(err, "failed to get AppMentionEvent")
		return
	}
	channelId := ev.Channel
	messageTs := ev.TimeStamp
	// init logger & context
	logger := c.log.WithValues("messageTs", messageTs)
	ctx := utils.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(err, "failed to initialize Slack client")
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong("None"))
		return
	}

	if err := sc.PostMessage(ctx, channelId, view.ShowVersion()); err != nil {
		logger.Error(err, "failed to post message")
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong("None"))
		return
	}
}
