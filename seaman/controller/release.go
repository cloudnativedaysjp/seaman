package controller

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/xerrors"

	"github.com/cloudnativedaysjp/seaman/seaman/api"
	"github.com/cloudnativedaysjp/seaman/seaman/infra/gitcommand"
	"github.com/cloudnativedaysjp/seaman/seaman/infra/githubapi"
	infra_slack "github.com/cloudnativedaysjp/seaman/seaman/infra/slack"
	"github.com/cloudnativedaysjp/seaman/seaman/service"
	"github.com/cloudnativedaysjp/seaman/seaman/utils"
	"github.com/cloudnativedaysjp/seaman/seaman/view"
)

type Target struct {
	Url        string
	BaseBranch string
}

type ReleaseController struct {
	slackFactory infra_slack.SlackClientFactory
	service      *service.ReleaseService
	log          logr.Logger

	targets []Target
}

func NewReleaseController(
	logger logr.Logger,
	slackFactory infra_slack.SlackClientFactory,
	gitcommand gitcommand.GitCommandClient,
	githubapi githubapi.GitHubApiClient,
	targets []Target,
) *ReleaseController {
	service := service.NewReleaseService(gitcommand, githubapi)
	return &ReleaseController{slackFactory, service, logger, targets}
}

func (c *ReleaseController) SelectRepository(evt *socketmode.Event, client *socketmode.Client) {
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
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *ReleaseController) SelectReleaseLevel(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)

	interaction, err := getInteractionCallback(evt)
	if err != nil {
		c.log.Error(err, "failed to get InteractionCallback")
		return
	}
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := getCallbackValueOnStaticSelect(interaction)
	// init logger & context
	logger := c.log.WithValues("messageTs", messageTs)
	ctx := utils.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to initialize Slack client")
		return
	}

	orgRepo, err := api.NewOrgRepo(callbackValue)
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

	interaction, err := getInteractionCallback(evt)
	if err != nil {
		c.log.Error(err, "failed to get InteractionCallback")
		return
	}
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := getCallbackValueOnStaticSelect(interaction)
	// init logger & context
	logger := c.log.WithValues("messageTs", messageTs)
	ctx := utils.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to initialize Slack client")
		return
	}

	orgRepoLevel, err := api.NewOrgRepoLevel(callbackValue)
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

	interaction, err := getInteractionCallback(evt)
	if err != nil {
		c.log.Error(err, "failed to get InteractionCallback")
		return
	}
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := getCallbackValueOnStaticSelect(interaction)
	// init logger & context
	logger := c.log.WithValues("messageTs", messageTs)
	ctx := utils.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to initialize Slack client")
		return
	}

	orgRepoLevel, err := api.NewOrgRepoLevel(callbackValue)
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
