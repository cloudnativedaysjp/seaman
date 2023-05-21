package controller

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/exp/slog"

	infra_slack "github.com/cloudnativedaysjp/seaman/internal/infra/slack"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/view"
	"github.com/cloudnativedaysjp/seaman/pkg/log"
)

type CommonController struct {
	slackFactory infra_slack.SlackClientFactory
	log          *slog.Logger
}

func NewCommonController(
	logger *slog.Logger,
	slackFactory infra_slack.SlackClientFactory,
) *CommonController {
	return &CommonController{slackFactory, logger}
}

func (c *CommonController) ShowCommands(ctx context.Context,
	ev *slackevents.AppMentionEvent,
	client *socketmode.Client, subcommands map[string]string,
) {
	logger := log.FromContext(ctx)
	channelId := ev.Channel
	messageTs := ev.TimeStamp
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize Slack client: %v", err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.PostMessage(ctx, channelId,
		view.ShowCommands(subcommands),
	); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *CommonController) ShowVersion(ctx context.Context, ev *slackevents.AppMentionEvent, client *socketmode.Client) {
	logger := log.FromContext(ctx)
	channelId := ev.Channel
	messageTs := ev.TimeStamp
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

func (c *CommonController) InteractionNothingToDo(ctx context.Context, interaction slack.InteractionCallback, client *socketmode.Client) {
}

func (c *CommonController) InteractionCancel(ctx context.Context, interaction slack.InteractionCallback, client *socketmode.Client) {
	logger := log.FromContext(ctx)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
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
