package controller

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/exp/slog"
	"golang.org/x/xerrors"

	infra_slack "github.com/cloudnativedaysjp/seaman/internal/infra/slack"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/view"
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
) error {
	channelId := ev.Channel
	messageTs := ev.TimeStamp
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to initialize Slack client: %w", err)
	}

	if err := sc.PostMessage(ctx, channelId,
		view.ShowCommands(subcommands),
	); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}
	return nil
}

func (c *CommonController) ShowVersion(ctx context.Context, ev *slackevents.AppMentionEvent, client *socketmode.Client) error {
	channelId := ev.Channel
	messageTs := ev.TimeStamp
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to initialize Slack client: %w", err)
	}

	if err := sc.PostMessage(ctx, channelId, view.ShowVersion()); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}
	return nil
}

func (c *CommonController) InteractionNothingToDo(ctx context.Context, interaction slack.InteractionCallback, client *socketmode.Client) error {
	return nil
}

func (c *CommonController) InteractionCancel(ctx context.Context, interaction slack.InteractionCallback, client *socketmode.Client) error {
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to initialize Slack client: %v", err)
	}

	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.Canceled()); err != nil {
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %v", err)
	}
	return nil
}
