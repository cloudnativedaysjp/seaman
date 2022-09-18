//go:build test_githubapi

package githubapi

import (
	"context"
	"os"
	"testing"
)

func Test_GitHubApiDriver(t *testing.T) {
	driver := NewGitHubApiDriver(os.Getenv("GITHUB_TOKEN"))

	t.Run(`HealthCheck`, func(t *testing.T) {
		err := driver.HealthCheck()
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
		prNum, err := driver.CreatePullRequest(ctx, org, repo, headBranch, baseBranch, "demo", "hoge\n`fuga`\n**piyo**")
		if err != nil {
			t.Fatalf("error: %s", err)
		}

		// LabelPullRequest
		if err := driver.LabelPullRequest(ctx, org, repo, prNum, label); err != nil {
			t.Fatalf("error: %s", err)
		}

		// MergePullRequest
		if err := driver.MergePullRequest(ctx, org, repo, prNum); err != nil {
			t.Fatalf("error: %s", err)
		}
	})
}
