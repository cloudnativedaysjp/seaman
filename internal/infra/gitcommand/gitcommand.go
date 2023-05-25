//go:generate go run github.com/golang/mock/mockgen -package mock -source=gitcommand.go -destination=mock/gitcommand.go

package gitcommand

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
)

type GitCommandClient interface {
	Clone(ctx context.Context, org, repo string, opt CloneOpt) (string, error)
	CommitAll(ctx context.Context, dirPath, commitMsg string) error
	CommitAllAmend(ctx context.Context, dirPath string) error
	HealthCheck() (err error)
	Push(ctx context.Context, dirPath string) error
	Remove(ctx context.Context, dirPath string) error
	Restore(ctx context.Context, dirPath, sourceBranch string, filePaths []string) error
	SwitchNewBranch(ctx context.Context, dirPath, branch string) error
}

type GitCommandClientImpl struct {
	user  string
	email string
	token string
}

func NewGitCommandClientImpl(user, token string) GitCommandClient {
	return &GitCommandClientImpl{user, fmt.Sprintf("%s@users.noreply.github.com", user), token}
}

var (
	baseURL = `https://%s:%s@github.com`
)

type CloneOpt struct {
	Branch string
	Depth  int
}

func (g *GitCommandClientImpl) Clone(ctx context.Context, org, repo string, opt CloneOpt) (string, error) {
	downloadDir := fmt.Sprintf("/tmp/%s", filepath.Base(repo))
	commands := []string{
		"git", "clone",
	}
	if opt.Branch != "" {
		commands = append(commands, "-b", opt.Branch)
	}
	if opt.Depth != 0 {
		commands = append(commands, "--depth", strconv.Itoa(opt.Depth))
	}
	commands = append(commands,
		strings.Join([]string{fmt.Sprintf(baseURL, g.user, g.token), org, repo}, "/"), // https://<user>:<token>@github.com/<org>/<repo>
		downloadDir,
	)
	cmd := exec.CommandContext(ctx, commands[0], commands[1:]...)
	stderr := cmd.Stderr
	if _, err := cmd.Output(); err != nil {
		return "", xerrors.Errorf("%v: %w", stderr, err)
	}
	return downloadDir, nil
}

func (g *GitCommandClientImpl) CommitAll(ctx context.Context, dirPath, commitMsg string) error {
	cmd := exec.CommandContext(ctx, "git", "config", "user.name", g.user)
	cmd.Dir = dirPath
	if _, err := cmd.Output(); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	cmd = exec.CommandContext(ctx, "git", "config", "user.email", g.email)
	cmd.Dir = dirPath
	stderr := cmd.Stderr
	if _, err := cmd.Output(); err != nil {
		return xerrors.Errorf("%v: %w", stderr, err)
	}
	cmd = exec.CommandContext(ctx, "git", "add", "-A")
	cmd.Dir = dirPath
	stderr = cmd.Stderr
	if _, err := cmd.Output(); err != nil {
		return xerrors.Errorf("%v: %w", stderr, err)
	}
	cmd = exec.CommandContext(ctx, "git", "commit", "--allow-empty", "-m", commitMsg)
	cmd.Dir = dirPath
	stderr = cmd.Stderr
	if _, err := cmd.Output(); err != nil {
		return xerrors.Errorf("%v: %w", stderr, err)
	}
	return nil
}

func (g *GitCommandClientImpl) CommitAllAmend(ctx context.Context, dirPath string) error {
	cmd := exec.CommandContext(ctx, "git", "config", "user.name", g.user)
	cmd.Dir = dirPath
	stderr := cmd.Stderr
	if _, err := cmd.Output(); err != nil {
		return xerrors.Errorf("%v: %w", stderr, err)
	}
	cmd = exec.CommandContext(ctx, "git", "config", "user.email", g.email)
	cmd.Dir = dirPath
	stderr = cmd.Stderr
	if _, err := cmd.Output(); err != nil {
		return xerrors.Errorf("%v: %w", stderr, err)
	}
	cmd = exec.CommandContext(ctx, "git", "add", "-A")
	cmd.Dir = dirPath
	stderr = cmd.Stderr
	if _, err := cmd.Output(); err != nil {
		return xerrors.Errorf("%v: %w", stderr, err)
	}
	cmd = exec.CommandContext(ctx,
		"git", "commit", "--amend", "--no-edit", "--author", fmt.Sprintf("%s <%s>", g.user, g.email))
	cmd.Dir = dirPath
	stderr = cmd.Stderr
	if _, err := cmd.Output(); err != nil {
		return xerrors.Errorf("%v: %w", stderr, err)
	}
	return nil
}

func (g *GitCommandClientImpl) HealthCheck() (err error) {
	return nil
}

func (g *GitCommandClientImpl) Push(ctx context.Context, dirPath string) error {
	cmd := exec.CommandContext(ctx, "git", "push", "origin", "HEAD")
	cmd.Dir = dirPath
	stderr := cmd.Stderr
	if _, err := cmd.Output(); err != nil {
		return xerrors.Errorf("%v: %w", stderr, err)
	}
	return nil
}

func (g *GitCommandClientImpl) Remove(ctx context.Context, dirPath string) error {
	if err := os.RemoveAll(dirPath); err != nil {
		return xerrors.Errorf("os.RemoveAll failed: %w", err)
	}
	return nil
}

func (g *GitCommandClientImpl) Restore(ctx context.Context, dirPath, sourceBranch string, filePaths []string) error {
	cmd := exec.CommandContext(ctx, "git", "restore", "-s", "origin/"+sourceBranch, strings.Join(filePaths, " "))
	cmd.Dir = dirPath
	stderr := cmd.Stderr
	if _, err := cmd.Output(); err != nil {
		return xerrors.Errorf("%v: %w", stderr, err)
	}

	return nil
}

func (g *GitCommandClientImpl) SwitchNewBranch(ctx context.Context, dirPath, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "switch", "-c", branch)
	cmd.Dir = dirPath
	stderr := cmd.Stderr
	if _, err := cmd.Output(); err != nil {
		return xerrors.Errorf("%v: %w", stderr, err)
	}
	return nil
}
