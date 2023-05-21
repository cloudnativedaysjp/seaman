//go:generate go run github.com/golang/mock/mockgen -package mock -source=github_release.go -destination=mock/github_release.go

package service

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/xerrors"

	"github.com/cloudnativedaysjp/seaman/internal/infra/gitcommand"
	"github.com/cloudnativedaysjp/seaman/internal/infra/githubapi"
	"github.com/cloudnativedaysjp/seaman/pkg/log"
)

type GitHubIface interface {
	CreatePullRequestWithEmptyCommit(ctx context.Context,
		org, repo, level string,
		targetBaseBranch string, headBranchSuffix string,
	) (prNum int, err error)

	SeparatePullRequests(ctx context.Context,
		org, repo string, prNum int, targetBaseBranch string, prBranch string,
	) (prNumDev, prNumProd int, err error)
}

type GitHub struct {
	gitcommand gitcommand.GitCommandClient
	githubapi  githubapi.GitHubApiClient
}

func NewGitHubService(
	gitcommand gitcommand.GitCommandClient,
	githubapi githubapi.GitHubApiClient,
) GitHubIface {
	return &GitHub{gitcommand, githubapi}
}

func (s *GitHub) CreatePullRequestWithEmptyCommit(ctx context.Context,
	org, repo, level string,
	targetBaseBranch string, headBranchSuffix string,
) (int, error) {
	const (
		emptyPrHeadBranchPrefix = "release/bot_"
	)
	logger := log.FromContext(ctx)
	headBranchName := emptyPrHeadBranchPrefix + headBranchSuffix

	//
	// clone repo to working dir
	//
	repoDir, err := s.gitcommand.Clone(ctx, org, repo, gitcommand.CloneOpt{Branch: targetBaseBranch, Depth: 1})
	if err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}
	// remove working dir finally
	defer func() {
		if err := s.gitcommand.Remove(ctx, repoDir); err != nil {
			logger.Warn(fmt.Sprintf("failed to remove working directory: %v", err))
		}
	}()

	//
	// switch -> empty commit -> push
	//
	if err := s.gitcommand.SwitchNewBranch(ctx, repoDir, headBranchName); err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}
	if err := s.gitcommand.CommitAll(ctx, repoDir, "[Bot] for release!!"); err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}
	if err := s.gitcommand.Push(ctx, repoDir); err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}

	//
	// create PR -> label
	//
	prNum, err := s.githubapi.CreatePullRequest(ctx, org, repo, headBranchName,
		targetBaseBranch, "[dreamkast-releasebot] Automatic Release", "Automatic Release")
	if err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}
	if err := s.githubapi.CreateLabels(ctx, org, repo, prNum, []string{level}); err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}

	return prNum, nil
}

func (s *GitHub) SeparatePullRequests(ctx context.Context,
	org, repo string, prNum int, targetBaseBranch string, prBranch string,
) (int, int, error) {
	title, changedFilepaths, err := s.githubapi.GetPullRequestTitleAndChangedFilepaths(ctx, org, repo, prNum)
	if err != nil {
		return 0, 0, xerrors.Errorf("githubapi.GetPullRequestChangedFilepaths failed: %w", err)
	}
	prNumDev, err := s.separatePullRequest(ctx,
		org, repo, targetBaseBranch, prBranch, title, changedFilepaths, "development")
	if err != nil {
		return 0, 0, xerrors.Errorf("separatePullRequest(dev) failed: %w", err)
	}
	prNumProd, err := s.separatePullRequest(ctx,
		org, repo, targetBaseBranch, prBranch, title, changedFilepaths, "production")
	if err != nil {
		return 0, 0, xerrors.Errorf("separatePullRequest(prod) failed: %w", err)
	}

	if err := s.githubapi.CreateIssueComment(ctx, org, repo, prNumProd, fmt.Sprintf(`Closes #%d`, prNum)); err != nil {
		return 0, 0, xerrors.Errorf("githubapi.CreateIssueComment failed: %w", err)
	}

	return prNumDev, prNumProd, nil
}

func (s *GitHub) separatePullRequest(ctx context.Context,
	org, repo string, targetBaseBranch string, prBranch string,
	title string, changedFilepaths []string, environment string,
) (int, error) {
	logger := log.FromContext(ctx)
	headBranch := fmt.Sprintf("%s_%s", prBranch, environment)

	//
	// clone repo to working dir
	//
	repoDir, err := s.gitcommand.Clone(ctx, org, repo, gitcommand.CloneOpt{Branch: prBranch})
	if err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}
	// remove working dir finally
	defer func() {
		if err := s.gitcommand.Remove(ctx, repoDir); err != nil {
			logger.Warn(fmt.Sprintf("failed to remove working directory: %v", err))
		}
	}()

	//
	// switch -> update files -> commit --amend -> push
	//
	if err := s.gitcommand.SwitchNewBranch(ctx, repoDir, headBranch); err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}
	if err := s.restoreFiles(ctx, repoDir, targetBaseBranch, changedFilepaths, environment); err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}
	if err := s.gitcommand.CommitAllAmend(ctx, repoDir); err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}
	if err := s.gitcommand.Push(ctx, repoDir); err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}

	//
	// create PR -> label
	//
	prNum, err := s.githubapi.CreatePullRequest(ctx, org, repo, headBranch,
		targetBaseBranch, title, "")
	if err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}
	if err := s.githubapi.CreateLabels(ctx, org, repo, prNum, []string{"dependencies"}); err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}

	return prNum, nil
}

func (s *GitHub) restoreFiles(ctx context.Context,
	repoDir string, sourceBranch string, changedFilepaths []string, environment string,
) error {
	fl := false
	restoredFilePaths := []string{}
	for _, fpath := range changedFilepaths {
		if strings.Contains(fpath, fmt.Sprintf("/%s/", environment)) {
			fl = true
		} else {
			restoredFilePaths = append(restoredFilePaths, fpath)
		}
	}
	if !fl {
		return xerrors.Errorf(`all of changedFilepaths don't contain "/%s/"`, environment)
	}
	if err := s.gitcommand.Restore(ctx, repoDir, sourceBranch, restoredFilePaths); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	return nil
}
