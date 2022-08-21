package controller

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	slack_driver "github.com/cloudnativedaysjp/slackbot/infrastructure/slack"
	"github.com/cloudnativedaysjp/slackbot/view"
)

type CommonController struct {
	slackFactory slack_driver.SlackDriverFactoryIface
	log          *zap.Logger

	subcommands []string
}

func NewCommonController(
	slackFactory slack_driver.SlackDriverFactoryIface,
	subcommands []string,
) *CommonController {
	logger, _ := zap.NewDevelopment()
	return &CommonController{slackFactory, logger, subcommands}
}

func (c *CommonController) ShowCommands(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	ctx := context.Background()
	// this handler is intended to be called by only incoming slackevents.AppMention.
	// So ignore validation of casting.
	ev := evt.Data.(slackevents.EventsAPIEvent).InnerEvent.Data.(*slackevents.AppMentionEvent)
	channelId := ev.Channel
	messageTs := ev.TimeStamp
	// new instances
	logger := c.log.With(zap.String("messageTs", messageTs)).Sugar()
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Errorf("failed to initialize Slack client: %v\n", err)
	}

	if err := sc.PostMessage(ctx, channelId,
		view.ShowCommands(c.subcommands),
	); err != nil {
		logger.Errorf("failed to post message: %v\n", err)
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong("None"))
	}
}

func (c *CommonController) InteractionCancel(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	ctx := context.Background()
	// this handler is intended to be called by only incoming slack.InteractionCallback.
	// So ignore validation of casting.
	interaction := evt.Data.(slack.InteractionCallback)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	// new instances
	logger := c.log.With(zap.String("messageTs", messageTs)).Sugar()
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Errorf("failed to initialize Slack client: %v\n", err)
	}

	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.Canceled()); err != nil {
		logger.Errorf("failed to post message: %v", err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
	}
}
