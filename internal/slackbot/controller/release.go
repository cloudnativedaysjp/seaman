package controller

import (
	"context"
	"fmt"

	"github.com/slack-go/slack/socketmode"
	"golang.org/x/exp/slog"

	"github.com/cloudnativedaysjp/seaman/internal/infra/gitcommand"
	"github.com/cloudnativedaysjp/seaman/internal/infra/githubapi"
	infra_slack "github.com/cloudnativedaysjp/seaman/internal/infra/slack"
	"github.com/cloudnativedaysjp/seaman/internal/log"
	"github.com/cloudnativedaysjp/seaman/internal/service"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/api"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/view"
)

type Target struct {
	Url        string
	BaseBranch string
}

type ReleaseController struct {
	slackFactory infra_slack.SlackClientFactory
	service      service.GitHubIface
	log          *slog.Logger

	targets []Target
}

func NewReleaseController(
	logger *slog.Logger,
	slackFactory infra_slack.SlackClientFactory,
	gitcommand gitcommand.GitCommandClient,
	githubapi githubapi.GitHubApiClient,
	targets []Target,
) *ReleaseController {
	service := service.NewGitHubService(gitcommand, githubapi)
	return &ReleaseController{slackFactory, service, logger, targets}
}

func (c *ReleaseController) SelectRepository(evt *socketmode.Event, client *socketmode.Client) {
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
	}

	var targetUrls []string
	for _, target := range c.targets {
		targetUrls = append(targetUrls, target.Url)
	}

	if err := sc.PostMessage(ctx, channelId,
		view.ReleaseListRepo(targetUrls),
	); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *ReleaseController) SelectReleaseLevel(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)

	interaction, err := getInteractionCallback(evt)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to get InteractionCallback: %v", err))
		return
	}
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := getCallbackValueOnStaticSelect(interaction)
	// init logger & context
	logger := c.log.With("messageTs", messageTs)
	ctx := log.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize Slack client: %v", err))
		return
	}

	orgRepo, err := api.NewOrgRepo(callbackValue)
	if err != nil {
		logger.Error(fmt.Sprintf("invalid callback value: %v", err),
			"callbackValue", callbackValue)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.ReleaseListLevel(orgRepo)); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *ReleaseController) SelectConfirmation(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)

	interaction, err := getInteractionCallback(evt)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to get InteractionCallback: %v", err))
		return
	}
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := getCallbackValueOnButton(interaction)
	// init logger & context
	logger := c.log.With("messageTs", messageTs)
	ctx := log.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize Slack client: %v", err))
		return
	}

	orgRepoLevel, err := api.NewOrgRepoLevel(callbackValue)
	if err != nil {
		logger.Error(fmt.Sprintf("invalid callback value: %v", err),
			"callbackValue", callbackValue)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	if err := sc.UpdateMessage(
		ctx, channelId, messageTs, view.ReleaseConfirmation(orgRepoLevel),
	); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *ReleaseController) CreatePullRequestForRelease(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)

	interaction, err := getInteractionCallback(evt)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to get InteractionCallback: %v", err))
		return
	}
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := getCallbackValueOnButton(interaction)
	// init logger & context
	logger := c.log.With("messageTs", messageTs)
	ctx := log.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize Slack client: %v", err))
		return
	}

	orgRepoLevel, err := api.NewOrgRepoLevel(callbackValue)
	if err != nil {
		logger.Error(fmt.Sprintf("invalid callback value: %v", err),
			"callbackValue", callbackValue)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.ReleaseProcessing()); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
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
	prNum, err := c.service.CreatePullRequestWithEmptyCommit(ctx,
		orgRepoLevel.Org(), orgRepoLevel.Repo(), orgRepoLevel.Level(), baseBranch, messageTs)
	if err != nil {
		logger.Error(fmt.Sprintf("service.CreatePullRequest was failed: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.UpdateMessage(
		ctx, channelId, messageTs, view.ReleaseDisplayPrLink(orgRepoLevel, prNum),
	); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}
