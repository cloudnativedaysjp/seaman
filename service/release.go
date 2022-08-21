package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/cloudnativedaysjp/slackbot/infrastructure/gitcommand"
	"github.com/cloudnativedaysjp/slackbot/infrastructure/githubapi"
	slack_driver "github.com/cloudnativedaysjp/slackbot/infrastructure/slack"
	"github.com/cloudnativedaysjp/slackbot/model"
	"github.com/cloudnativedaysjp/slackbot/view"
)

const (
	releaseHeadBranchPrefix = "release/bot_"
)

type ReleaseService struct {
	gitcommand gitcommand.GitCommandIface
	githubapi  githubapi.GitHubApiIface
	log        *zap.Logger
}

func NewReleaseService(
	gitcommand gitcommand.GitCommandIface,
	githubapi githubapi.GitHubApiIface,
) *ReleaseService {
	logger, _ := zap.NewDevelopment()
	return &ReleaseService{gitcommand, githubapi, logger}
}

func (s ReleaseService) CreatePullRequest(ctx context.Context,
	sc slack_driver.SlackIface, channelId, messageTs string,
	orgRepoLevel model.OrgRepoLevel, targetBaseBranch string,
) {
	logger := s.log.With(zap.String("messageTs", messageTs)).Sugar()
	org := orgRepoLevel.Org()
	repo := orgRepoLevel.Repo()
	level := orgRepoLevel.Level()
	releaseHeadBranch := releaseHeadBranchPrefix + messageTs

	// clone repo to working dir
	repoDir, err := s.gitcommand.Clone(ctx, org, repo)
	if err != nil {
		logger.Error(err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	// remove working dir finally
	defer func() {
		if err := s.gitcommand.Remove(ctx, repoDir); err != nil {
			logger.Error(err)
			return
		}
	}()
	// switch -> empty commit -> push
	if err := s.gitcommand.SwitchNewBranch(ctx, repoDir, releaseHeadBranch); err != nil {
		logger.Errorf("SwitchNewBranch() was failed: %v", err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	if err := s.gitcommand.CommitAll(ctx, repoDir, "[Bot] for release!!"); err != nil {
		logger.Errorf("CommitAll() was failed: %v", err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	if err := s.gitcommand.Push(ctx, repoDir); err != nil {
		logger.Errorf("Push() was failed: %v", err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
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
		logger.Errorf("CreatePullRequest() was failed: %v", err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	if err := s.githubapi.LabelPullRequest(ctx, org, repo, prNum, level); err != nil {
		logger.Errorf("LabelPullRequest() was failed: %v", err)
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
	passFlag = true
	// update Slack Message
	if err := sc.UpdateMessage(
		ctx, channelId, messageTs, view.ReleaseDisplayPrLink(orgRepoLevel, prNum),
	); err != nil {
		logger.Errorf("UpdateMessage() was failed: %v", err)
		return
	}
}
