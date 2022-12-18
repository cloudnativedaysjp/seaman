//go:generate go run github.com/golang/mock/mockgen -package mock -source=gitcommand.go -destination=mock/gitcommand.go

package gitcommand

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/xerrors"
)

type GitCommandClient interface {
	HealthCheck() (err error)
	Clone(ctx context.Context, org, repo string) (string, error)
	SwitchNewBranch(ctx context.Context, dirPath, branch string) error
	CommitAll(ctx context.Context, dirPath, commitMsg string) error
	Push(ctx context.Context, dirPath string) error
	Remove(ctx context.Context, dir string) error
}

type GitCommandClientImpl struct {
	user  string
	email string
	token string
}

func NewGitCommandClientImpl(user, token string) *GitCommandClientImpl {
	return &GitCommandClientImpl{user, fmt.Sprintf("%s@users.noreply.github.com", user), token}
}

var (
	baseURL = `https://%s:%s@github.com`
)

func (g *GitCommandClientImpl) HealthCheck() (err error) {
	return nil
}

func (g *GitCommandClientImpl) Clone(ctx context.Context, org, repo string) (string, error) {
	downloadDir := fmt.Sprintf("/tmp/%s", filepath.Base(repo))
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1",
		strings.Join([]string{fmt.Sprintf(baseURL, g.user, g.token), org, repo}, "/"), // https://<user>:<token>@github.com/<org>/<repo>
		downloadDir)
	if msg, err := cmd.CombinedOutput(); err != nil {
		return "", xerrors.Errorf("message: %w", msg)
	}
	return downloadDir, nil
}

func (g *GitCommandClientImpl) SwitchNewBranch(ctx context.Context, dirPath, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "switch", "-c", branch)
	cmd.Dir = dirPath
	if msg, err := cmd.CombinedOutput(); err != nil {
		return xerrors.Errorf("message: %w", msg)
	}
	return nil
}

func (g *GitCommandClientImpl) CommitAll(ctx context.Context, dirPath, commitMsg string) error {
	cmd := exec.CommandContext(ctx, "git", "config", "user.name", g.user)
	cmd.Dir = dirPath
	if msg, err := cmd.CombinedOutput(); err != nil {
		return xerrors.Errorf("message: %w", msg)
	}
	cmd = exec.CommandContext(ctx, "git", "config", "user.email", g.email)
	cmd.Dir = dirPath
	if msg, err := cmd.CombinedOutput(); err != nil {
		return xerrors.Errorf("message: %w", msg)
	}
	cmd = exec.CommandContext(ctx, "git", "add", "-A")
	cmd.Dir = dirPath
	if msg, err := cmd.CombinedOutput(); err != nil {
		return xerrors.Errorf("message: %w", msg)
	}
	cmd = exec.CommandContext(ctx, "git", "commit", "--allow-empty", "-m", commitMsg)
	cmd.Dir = dirPath
	if msg, err := cmd.CombinedOutput(); err != nil {
		return xerrors.Errorf("message: %w", msg)
	}
	return nil
}

func (g *GitCommandClientImpl) Push(ctx context.Context, dirPath string) error {
	cmd := exec.CommandContext(ctx, "git", "push", "origin", "HEAD")
	cmd.Dir = dirPath
	if msg, err := cmd.CombinedOutput(); err != nil {
		return xerrors.Errorf("message: %w", msg)
	}
	return nil
}

func (g *GitCommandClientImpl) Remove(ctx context.Context, dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	return nil
}
