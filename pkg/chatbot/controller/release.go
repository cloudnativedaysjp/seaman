package controller

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/cloudnativedaysjp/chatbot/pkg/chatbot/dto"
	"github.com/cloudnativedaysjp/chatbot/pkg/chatbot/service"
	"github.com/cloudnativedaysjp/chatbot/pkg/chatbot/view"
	"github.com/cloudnativedaysjp/chatbot/pkg/gitcommand"
	"github.com/cloudnativedaysjp/chatbot/pkg/githubapi"
	slack_driver "github.com/cloudnativedaysjp/chatbot/pkg/slack"
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
	logger *zap.Logger,
	slackFactory slack_driver.SlackDriverFactoryIface,
	gitcommand gitcommand.GitCommandIface,
	githubapi githubapi.GitHubApiIface,
	targets []Target,
) *ReleaseController {
	service := service.NewReleaseService(gitcommand, githubapi)
	return &ReleaseController{slackFactory, service, logger, targets}
}

func (c *ReleaseController) SelectRepository(evt *socketmode.Event, client *socketmode.Client) {
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
		logger.Error(xerrors.Errorf("message: %w", err), "failed to initialize Slack client")
	}

	var targetUrls []string
	for _, target := range c.targets {
		targetUrls = append(targetUrls, target.Url)
	}

	if err := sc.PostMessage(ctx, channelId,
		view.ReleaseListRepo(targetUrls),
	); err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to post message")
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong("None"))
	}
}

func (c *ReleaseController) SelectReleaseLevel(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	// this handler is intended to be called by only incoming slack.InteractionCallback.
	// So ignore validation of casting.
	interaction := evt.Data.(slack.InteractionCallback)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := interaction.ActionCallback.BlockActions[0].SelectedOption.Value
	// init logger & context
	logger := zapr.NewLogger(c.log.With(zap.String("messageTs", messageTs)))
	ctx := logr.NewContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to initialize Slack client")
	}

	orgRepo, err := dto.NewOrgRepo(callbackValue)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "invalid callback value",
			"callbackValue", interaction.ActionCallback.BlockActions[0].Value)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.ReleaseListLevel(orgRepo)); err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to post message: %v", err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *ReleaseController) SelectConfirmation(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	// this handler is intended to be called by only incoming slack.InteractionCallback.
	// So ignore validation of casting.
	interaction := evt.Data.(slack.InteractionCallback)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := interaction.ActionCallback.BlockActions[0].Value
	// init logger & context
	logger := zapr.NewLogger(c.log.With(zap.String("messageTs", messageTs)))
	ctx := logr.NewContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to initialize Slack client")
		return
	}

	orgRepoLevel, err := dto.NewOrgRepoLevel(callbackValue)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "invalid callback value",
			"callbackValue", interaction.ActionCallback.BlockActions[0].Value)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	if err := sc.UpdateMessage(
		ctx, channelId, messageTs, view.ReleaseConfirmation(orgRepoLevel),
	); err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to post message")
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *ReleaseController) CreatePullRequestForRelease(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	// this handler is intended to be called by only incoming slack.InteractionCallback.
	// So ignore validation of casting.
	interaction := evt.Data.(slack.InteractionCallback)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := interaction.ActionCallback.BlockActions[0].Value
	// init logger & context
	logger := zapr.NewLogger(c.log.With(zap.String("messageTs", messageTs)))
	ctx := logr.NewContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to initialize Slack client")
		return
	}

	orgRepoLevel, err := dto.NewOrgRepoLevel(callbackValue)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "invalid callback value",
			"callbackValue", interaction.ActionCallback.BlockActions[0].Value)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.ReleaseProcessing()); err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to post message")
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
	if err := c.service.CreatePullRequest(ctx, sc, channelId, messageTs, orgRepoLevel, baseBranch); err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "service.CreatePullRequest was failed")
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}
