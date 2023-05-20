package controller

import (
	"context"
	"fmt"

	"github.com/slack-go/slack/socketmode"
	"golang.org/x/exp/slog"

	infra_slack "github.com/cloudnativedaysjp/seaman/internal/infra/slack"
	"github.com/cloudnativedaysjp/seaman/internal/log"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/view"
)

type CommonController struct {
	slackFactory infra_slack.SlackClientFactory
	log          *slog.Logger

	subcommands map[string]string
}

func NewCommonController(
	logger *slog.Logger,
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
		c.log.Error(fmt.Sprintf("failed to get AppMentionEvent: %v", err))
		return
	}
	channelId := ev.Channel
	messageTs := ev.TimeStamp
	// init logger & context
	logger := c.log.With("messageTs", messageTs)
	ctx := log.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize Slack client: %v", err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.PostMessage(ctx, channelId,
		view.ShowCommands(c.subcommands),
	); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *CommonController) ShowVersion(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)

	ev, err := getAppMentionEvent(evt)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to get AppMentionEvent: %v", err))
		return
	}
	channelId := ev.Channel
	messageTs := ev.TimeStamp
	// init logger & context
	logger := c.log.With("messageTs", messageTs)
	ctx := log.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize Slack client: %v", err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.PostMessage(ctx, channelId, view.ShowVersion()); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *CommonController) InteractionCancel(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)

	interaction, err := getInteractionCallback(evt)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to get InteractionCallback: %v", err))
		return
	}
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	// init logger & context
	logger := c.log.With("messageTs", messageTs)
	ctx := log.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize Slack client: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.Canceled()); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}
