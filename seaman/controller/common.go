package controller

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	infra_slack "github.com/cloudnativedaysjp/seaman/seaman/infrastructure/slack"
	"github.com/cloudnativedaysjp/seaman/seaman/view"
)

type CommonController struct {
	slackFactory infra_slack.SlackDriverFactoryIface
	log          *zap.Logger

	subcommands map[string]string
}

func NewCommonController(
	logger *zap.Logger,
	slackFactory infra_slack.SlackDriverFactoryIface,
	subcommands map[string]string,
) *CommonController {
	return &CommonController{slackFactory, logger, subcommands}
}

func (c *CommonController) ShowCommands(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	// this handler is intended to be called by only incoming slackevents.AppMention.
	// So ignore validation of casting.
	ev := evt.Data.(slackevents.EventsAPIEvent).InnerEvent.Data.(*slackevents.AppMentionEvent)
	channelId := ev.Channel
	messageTs := ev.TimeStamp
	// init logger & context
	logger := zapr.NewLogger(c.log.With(zap.String("messageTs", messageTs)))
	ctx := logr.NewContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(err, "failed to initialize Slack client")
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong("None"))
		return
	}

	if err := sc.PostMessage(ctx, channelId,
		view.ShowCommands(c.subcommands),
	); err != nil {
		logger.Error(err, "failed to post message")
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong("None"))
		return
	}
}

func (c *CommonController) InteractionCancel(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	// this handler is intended to be called by only incoming slack.InteractionCallback.
	// So ignore validation of casting.
	interaction := evt.Data.(slack.InteractionCallback)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	// init logger & context
	logger := zapr.NewLogger(c.log.With(zap.String("messageTs", messageTs)))
	ctx := logr.NewContext(context.Background(), logger)
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
