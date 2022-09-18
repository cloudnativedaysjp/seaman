package service

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/cloudnativedaysjp/seaman/seaman/dto"
	"github.com/cloudnativedaysjp/seaman/seaman/infrastructure/gitcommand"
	"github.com/cloudnativedaysjp/seaman/seaman/infrastructure/githubapi"
	infra_slack "github.com/cloudnativedaysjp/seaman/seaman/infrastructure/slack"
	"github.com/cloudnativedaysjp/seaman/seaman/view"
)

const (
	releaseHeadBranchPrefix = "release/bot_"
)

type ReleaseService struct {
	gitcommand gitcommand.GitCommandIface
	githubapi  githubapi.GitHubApiIface
}

func NewReleaseService(
	gitcommand gitcommand.GitCommandIface,
	githubapi githubapi.GitHubApiIface,
) *ReleaseService {
	return &ReleaseService{gitcommand, githubapi}
}

func (s ReleaseService) CreatePullRequest(ctx context.Context,
	sc infra_slack.SlackIface, channelId, messageTs string,
	orgRepoLevel dto.OrgRepoLevel, targetBaseBranch string,
) error {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		zaplogger, _ := zap.NewDevelopment()
		logger = zapr.NewLogger(zaplogger)
	}
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