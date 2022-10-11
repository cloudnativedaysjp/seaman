package controller

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/slack-go/slack/socketmode"

	infra_slack "github.com/cloudnativedaysjp/seaman/seaman/infra/slack"
	"github.com/cloudnativedaysjp/seaman/seaman/utils"
	"github.com/cloudnativedaysjp/seaman/seaman/view"
)

type CommonController struct {
	slackFactory infra_slack.SlackClientFactory
	log          logr.Logger

	subcommands map[string]string
}

func NewCommonController(
	logger logr.Logger,
	slackFactory infra_slack.SlackClientFactory,
	subcommands map[string]string,
) *CommonController {
	return &CommonController{slackFactory, logger, subcommands}
}

func (c *CommonController) NothingToDo(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
}

func (c *CommonController) ShowCommands(evt *socketmode.Event, client *socketmode.Client) {
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
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.PostMessage(ctx, channelId,
		view.ShowCommands(c.subcommands),
	); err != nil {
		logger.Error(err, "failed to post message")
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *CommonController) ShowVersion(evt *socketmode.Event, client *socketmode.Client) {
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
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.PostMessage(ctx, channelId, view.ShowVersion()); err != nil {
		logger.Error(err, "failed to post message")
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *CommonController) InteractionCancel(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)

	interaction, err := getInteractionCallback(evt)
	if err != nil {
		c.log.Error(err, "failed to get InteractionCallback")
		return
	}
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	// init logger & context
	logger := c.log.WithValues("messageTs", messageTs)
	ctx := utils.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(err, "failed to initialize Slack client")
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.Canceled()); err != nil {
		logger.Error(err, "failed to post message")
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}
