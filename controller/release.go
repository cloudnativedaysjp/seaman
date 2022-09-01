package controller

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"github.com/cloudnativedaysjp/chatbot/infrastructure/gitcommand"
	"github.com/cloudnativedaysjp/chatbot/infrastructure/githubapi"
	slack_driver "github.com/cloudnativedaysjp/chatbot/infrastructure/slack"
	"github.com/cloudnativedaysjp/chatbot/model"
	"github.com/cloudnativedaysjp/chatbot/service"
	"github.com/cloudnativedaysjp/chatbot/view"
)

type Target struct {
	Url        string
	BaseBranch string
}

type ReleaseController struct {
	slackFactory slack_driver.SlackDriverFactoryIface
	service      *service.ReleaseService
	log          *zap.Logger

	targets []Target
}

func NewReleaseController(
	slackFactory slack_driver.SlackDriverFactoryIface,
	gitcommand gitcommand.GitCommandIface,
	githubapi githubapi.GitHubApiIface,
	targets []Target,
) *ReleaseController {
	service := service.NewReleaseService(gitcommand, githubapi)
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

	var targetUrls []string
	for _, target := range c.targets {
		targetUrls = append(targetUrls, target.Url)
	}

	if err := sc.PostMessage(ctx, channelId,
		view.ReleaseListRepo(targetUrls),
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
		logger.Errorf(
			"ERROR: callback value is %s\n", interaction.ActionCallback.BlockActions[0].Value)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.ReleaseProcessing()); err != nil {
		logger.Errorf("failed to post message: %v", err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}

	var baseBranch string
	repoUrl := orgRepoLevel.RepositoryUrl()
	for _, target := range c.targets {
		if target.Url == repoUrl {
			baseBranch = target.BaseBranch
			break
		}
	}
	c.service.CreatePullRequest(ctx, sc, channelId, messageTs, orgRepoLevel, baseBranch)
}
