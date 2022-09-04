package gitcommand

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitCommandIface interface {
	HealthCheck() (err error)
	Clone(ctx context.Context, org, repo string) (string, error)
	SwitchNewBranch(ctx context.Context, dirPath, branch string) error
	CommitAll(ctx context.Context, dirPath, commitMsg string) error
	Push(ctx context.Context, dirPath string) error
	Remove(ctx context.Context, dir string) error
}

type GitCommandDriver struct {
	user  string
	email string
	token string
}

func NewGitCommandDriver(user, token string) *GitCommandDriver {
	return &GitCommandDriver{user, fmt.Sprintf("%s@users.noreply.github.com", user), token}
}

var (
	baseURL = `https://%s:%s@github.com`
)

func (g *GitCommandDriver) HealthCheck() (err error) {
	return nil
}

func (g *GitCommandDriver) Clone(ctx context.Context, org, repo string) (string, error) {
	downloadDir := fmt.Sprintf("/tmp/%s", filepath.Base(repo))
	cmd := exec.CommandContext(ctx, "git", "clone",
		strings.Join([]string{fmt.Sprintf(baseURL, g.user, g.token), org, repo}, "/"), // https://<user>:<token>@github.com/<org>/<repo>
		downloadDir)
	if out, err := cmd.Output(); err != nil {
		return "", fmt.Errorf(`Error: %v: %v`, err, out)
	}
	return downloadDir, nil
}

func (g *GitCommandDriver) SwitchNewBranch(ctx context.Context, dirPath, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "switch", "-c", branch)
	cmd.Dir = dirPath
	if out, err := cmd.Output(); err != nil {
		return fmt.Errorf(`Error: %v: %v`, err, out)
	}
	return nil
}

func (g *GitCommandDriver) CommitAll(ctx context.Context, dirPath, commitMsg string) error {
	cmd := exec.CommandContext(ctx, "git", "config", "user.name", g.user)
	cmd.Dir = dirPath
	if out, err := cmd.Output(); err != nil {
		return fmt.Errorf(`Error: %v: %v`, err, out)
	}
	cmd = exec.CommandContext(ctx, "git", "config", "user.email", g.email)
	cmd.Dir = dirPath
	if out, err := cmd.Output(); err != nil {
		return fmt.Errorf(`Error: %v: %v`, err, out)
	}
	cmd = exec.CommandContext(ctx, "git", "add", "-A")
	cmd.Dir = dirPath
	if out, err := cmd.Output(); err != nil {
		return fmt.Errorf(`Error: %v: %v`, err, out)
	}
	cmd = exec.CommandContext(ctx, "git", "commit", "--allow-empty", "-m", commitMsg)
	cmd.Dir = dirPath
	if out, err := cmd.Output(); err != nil {
		return fmt.Errorf(`Error: %v: %v`, err, out)
	}
	return nil
}

func (g *GitCommandDriver) Push(ctx context.Context, dirPath string) error {
	cmd := exec.CommandContext(ctx, "git", "push", "origin", "HEAD")
	cmd.Dir = dirPath
	if out, err := cmd.Output(); err != nil {
		return fmt.Errorf(`Error: %v: %v`, err, out)
	}
	return nil
}

func (g *GitCommandDriver) Remove(ctx context.Context, dir string) error {
	return os.RemoveAll(dir)
}
