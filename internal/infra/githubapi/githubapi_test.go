//go:build test_githubapi

package githubapi

import (
	"context"
	"os"
	"testing"
)

func Test_GitHubApiDriver(t *testing.T) {
	c := NewGitHubApiClient(os.Getenv("GITHUB_TOKEN"))

	t.Run(`HealthCheck`, func(t *testing.T) {
		err := c.HealthCheck()
		if err != nil {
			t.Fatalf("error: %s", err)
		}
	})

	t.Run(`CreatePullRequest & MergePullRequest`, func(t *testing.T) {
		ctx := context.Background()
		org := "ShotaKitazawa"
		repo := "dotfiles"
		headBranch := "demo"
		baseBranch := "master"
		label := "bug"

		// CreatePullRequest
		prNum, err := c.CreatePullRequest(ctx, org, repo, headBranch, baseBranch, "demo", "hoge\n`fuga`\n**piyo**")
		if err != nil {
			t.Fatalf("error: %s", err)
		}

		// LabelPullRequest
		if err := c.LabelPullRequest(ctx, org, repo, prNum, label); err != nil {
			t.Fatalf("error: %s", err)
		}
	})
}
