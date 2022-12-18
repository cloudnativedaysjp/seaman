package service

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"

	"github.com/cloudnativedaysjp/seaman/seaman/api"
	"github.com/cloudnativedaysjp/seaman/seaman/infra/gitcommand"
	"github.com/cloudnativedaysjp/seaman/seaman/infra/githubapi"
	infra_slack "github.com/cloudnativedaysjp/seaman/seaman/infra/slack"
	"github.com/cloudnativedaysjp/seaman/seaman/utils"
	"github.com/cloudnativedaysjp/seaman/seaman/view"
)

const (
	releaseHeadBranchPrefix = "bot/release/"
)

type ReleaseService struct {
	gitcommand gitcommand.GitCommandClient
	githubapi  githubapi.GitHubApiClient
}

func NewReleaseService(
	gitcommand gitcommand.GitCommandClient,
	githubapi githubapi.GitHubApiClient,
) *ReleaseService {
	return &ReleaseService{gitcommand, githubapi}
}

func (s ReleaseService) CreatePullRequest(ctx context.Context,
	sc infra_slack.SlackClient, channelId, messageTs string,
	orgRepoLevel api.OrgRepoLevel, targetBaseBranch string,
) error {
	logger := utils.FromContext(ctx)

	org := orgRepoLevel.Org()
	repo := orgRepoLevel.Repo()
	level := orgRepoLevel.Level()
	releaseHeadBranch := releaseHeadBranchPrefix + messageTs

	// clone repo to working dir
	repoDir, err := s.gitcommand.Clone(ctx, org, repo)
	if err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	// remove working dir finally
	defer func() {
		if err := s.gitcommand.Remove(ctx, repoDir); err != nil {
			logger.Info(fmt.Sprintf("failed to remove working directory: %v", err))
		}
	}()
	// switch -> empty commit -> push
	if err := s.gitcommand.SwitchNewBranch(ctx, repoDir, releaseHeadBranch); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	if err := s.gitcommand.CommitAll(ctx, repoDir, "[Bot] for release!!"); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	if err := s.gitcommand.Push(ctx, repoDir); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	passFlag := false
	defer func() {
		if !passFlag {
			if err := s.githubapi.DeleteBranch(ctx, org, repo, releaseHeadBranch); err != nil {
				logger.Info(fmt.Sprintf(
					"failed to remove remote branch (%s): %v", releaseHeadBranch, err))
			}
		}
	}()
	// create -> label -> merge PullRequest
	prNum, err := s.githubapi.CreatePullRequest(ctx, org, repo, releaseHeadBranch,
		targetBaseBranch, "[dreamkast-releasebot] Automatic Release", "Automatic Release")
	if err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	if err := s.githubapi.LabelPullRequest(ctx, org, repo, prNum, level); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	passFlag = true
	// update Slack Message
	if err := sc.UpdateMessage(
		ctx, channelId, messageTs, view.ReleaseDisplayPrLink(orgRepoLevel, prNum),
	); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	return nil
}
