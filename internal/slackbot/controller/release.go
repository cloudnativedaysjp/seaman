package controller

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/exp/slog"
	"golang.org/x/xerrors"

	"github.com/cloudnativedaysjp/seaman/internal/infra/gitcommand"
	"github.com/cloudnativedaysjp/seaman/internal/infra/githubapi"
	infra_slack "github.com/cloudnativedaysjp/seaman/internal/infra/slack"
	"github.com/cloudnativedaysjp/seaman/internal/service"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/api"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/view"
	"github.com/cloudnativedaysjp/seaman/pkg/log"
	"github.com/cloudnativedaysjp/seaman/pkg/utils"
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

func (c *ReleaseController) SelectRepository(ctx context.Context, ev *slackevents.AppMentionEvent, client *socketmode.Client) error {
	channelId := ev.Channel
	messageTs := ev.TimeStamp
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		return xerrors.Errorf("failed to initialize Slack client: %w", err)
	}

	var targetUrls []string
	for _, target := range c.targets {
		targetUrls = append(targetUrls, target.Url)
	}

	if err := sc.PostMessage(ctx, channelId,
		view.ReleaseListRepo(targetUrls),
	); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}
	return nil
}

//nolint:dupl
func (c *ReleaseController) SelectReleaseLevel(ctx context.Context, interaction slack.InteractionCallback, client *socketmode.Client) error {
	logger := log.FromContext(ctx)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs

	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		return xerrors.Errorf("failed to initialize Slack client: %w", err)
	}

	orgRepo, err := api.NewOrgRepo(utils.GetCallbackValueOnStaticSelect(interaction))
	if err != nil {
		logger.Debug(fmt.Sprintf("invalid callback value: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return nil
	}
	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.ReleaseListLevel(orgRepo)); err != nil {
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}
	return nil
}

//nolint:dupl
func (c *ReleaseController) SelectConfirmation(ctx context.Context, interaction slack.InteractionCallback, client *socketmode.Client) error {
	logger := log.FromContext(ctx)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		return xerrors.Errorf("failed to initialize Slack client: %w", err)
	}

	orgRepoLevel, err := api.NewOrgRepoLevel(utils.GetCallbackValueOnButton(interaction))
	if err != nil {
		logger.Debug(fmt.Sprintf("invalid callback value: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return nil
	}
	if err := sc.UpdateMessage(
		ctx, channelId, messageTs, view.ReleaseConfirmation(orgRepoLevel),
	); err != nil {
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}
	return nil
}

func (c *ReleaseController) CreatePullRequestForRelease(ctx context.Context, interaction slack.InteractionCallback, client *socketmode.Client) error {
	logger := log.FromContext(ctx)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs

	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		return xerrors.Errorf("failed to initialize Slack client: %w", err)
	}

	orgRepoLevel, err := api.NewOrgRepoLevel(utils.GetCallbackValueOnButton(interaction))
	if err != nil {
		logger.Debug(fmt.Sprintf("invalid callback value: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return nil
	}

	if err := sc.UpdateMessage(ctx, channelId, messageTs, view.ReleaseProcessing()); err != nil {
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
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
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("service.CreatePullRequest failed: %w", err)
	}

	if err := sc.UpdateMessage(
		ctx, channelId, messageTs, view.ReleaseDisplayPrLink(orgRepoLevel, prNum),
	); err != nil {
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}
	return nil
}
