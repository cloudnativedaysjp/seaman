package controller

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"github.com/cloudnativedaysjp/slackbot/infrastructure/gitcommand"
	"github.com/cloudnativedaysjp/slackbot/infrastructure/githubapi"
	slack_driver "github.com/cloudnativedaysjp/slackbot/infrastructure/slack"
	"github.com/cloudnativedaysjp/slackbot/model"
	"github.com/cloudnativedaysjp/slackbot/service"
	"github.com/cloudnativedaysjp/slackbot/view"
)

type ReleaseController struct {
	slackFactory slack_driver.SlackDriverFactoryIface
	service      *service.ReleaseService
	log          *zap.Logger

	targets []string
}

func NewReleaseController(
	slackFactory slack_driver.SlackDriverFactoryIface,
	gitcommand gitcommand.GitCommandIface,
	githubapi githubapi.GitHubApiIface,
	targets []string, baseBranch string,
) *ReleaseController {
	service := service.NewReleaseService(gitcommand, githubapi, baseBranch)
	logger, _ := zap.NewDevelopment()
	return &ReleaseController{slackFactory, service, logger, targets}
}

func (c *ReleaseController) SelectRepository(evt *socketmode.Event, client *socketmode.Client) {
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
		view.ReleaseListRepo(c.targets),
	); err != nil {
		logger.Errorf("failed to post message: %v\n", err)
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong("None"))
	}
}

func (c *ReleaseController) SelectReleaseLevel(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	ctx := context.Background()
	// this handler is intended to be called by only incoming slack.InteractionCallback.
	// So ignore validation of casting.
	interaction := evt.Data.(slack.InteractionCallback)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := interaction.ActionCallback.BlockActions[0].SelectedOption.Value
	// new instances
	logger := c.log.With(zap.String("messageTs", messageTs)).Sugar()
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Errorf("failed to initialize Slack client: %v\n", err)
	}

	orgRepo, err := model.NewOrgRepo(callbackValue)
	if err != nil {
		logger.Errorf("ERROR: callback value is %s\n", interaction.ActionCallback.BlockActions[0].Value)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.ReleaseListLevel(orgRepo)); err != nil {
		logger.Errorf("failed to post message: %v", err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *ReleaseController) SelectConfirmation(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	ctx := context.Background()
	// this handler is intended to be called by only incoming slack.InteractionCallback.
	// So ignore validation of casting.
	interaction := evt.Data.(slack.InteractionCallback)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := interaction.ActionCallback.BlockActions[0].Value
	// new instances
	logger := c.log.With(zap.String("messageTs", messageTs)).Sugar()
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Errorf("failed to initialize Slack client: %v\n", err)
		return
	}

	orgRepoLevel, err := model.NewOrgRepoLevel(callbackValue)
	if err != nil {
		logger.Errorf("ERROR: callback value is %s\n", interaction.ActionCallback.BlockActions[0].Value)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	if err := sc.UpdateMessage(
		ctx, channelId, messageTs, view.ReleaseConfirmation(orgRepoLevel),
	); err != nil {
		logger.Errorf("failed to post message: %v", err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *ReleaseController) CreatePullRequestForRelease(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	ctx := context.Background()
	// this handler is intended to be called by only incoming slack.InteractionCallback.
	// So ignore validation of casting.
	interaction := evt.Data.(slack.InteractionCallback)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := interaction.ActionCallback.BlockActions[0].Value
	// new instances
	logger := c.log.With(zap.String("messageTs", messageTs)).Sugar()
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Errorf("failed to initialize Slack client: %v\n", err)
		return
	}

	orgRepoLevel, err := model.NewOrgRepoLevel(callbackValue)
	if err != nil {
		logger.Errorf("ERROR: callback value is %s\n", interaction.ActionCallback.BlockActions[0].Value)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.ReleaseProcessing()); err != nil {
		logger.Errorf("failed to post message: %v", err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}

	c.service.CreatePullRequest(ctx, sc, channelId, messageTs, orgRepoLevel)
}
