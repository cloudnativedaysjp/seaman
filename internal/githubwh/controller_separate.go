package githubwh

import (
	"context"
	"fmt"

	"github.com/go-playground/webhooks/v6/github"

	"github.com/cloudnativedaysjp/seaman/pkg/log"
	"github.com/cloudnativedaysjp/seaman/pkg/utils"
)

func (c Controller) CommandSeparate(ctx context.Context, payload github.IssueCommentPayload, args []string) error {
	var (
		supportedRepos = []string{"cloudnativedaysjp/dreamkast-infra"}
		targetBranch   = "main"
		org            = payload.Repository.Owner.Login
		repo           = payload.Repository.Name
		prNum          = int(payload.Issue.Number)
	)
	logger := log.FromContext(ctx, "command", "/SEPARATE",
		"repo", payload.Repository.FullName,
		"number", payload.Issue.Number,
		"url", payload.Issue.URL,
	)
	if !utils.Contains(supportedRepos, payload.Repository.FullName) {
		logger.Info("unsupported repository")
		return nil
	}

	// Validate PullRequest
	validPr, headBranchName, err := c.githubapi.CheckPrIsForInfraAndCreatedByRenovate(ctx, org, repo, prNum)
	if err != nil {
		return err
	}
	if !validPr {
		logger.Info("unsupported pullRequest")
		return nil
	}

	// Separate PullRequest
	prNumDev, prNumProd, err := c.service.SeparatePullRequests(ctx, org, repo, prNum, targetBranch, headBranchName)
	if err != nil {
		return err
	}

	// Label "DO NOT MERGE"
	if err := c.githubapi.CreateLabels(ctx,
		org, repo, prNum, []string{"dependencies", "DO NOT MERGE"}); err != nil {
		return err
	}

	body := fmt.Sprintf(`
separated to the following PRs
* #%d
* #%d
**Please merge them instead of this PR.**
`, prNumDev, prNumProd)

	return c.githubapi.CreateIssueComment(ctx, org, repo, prNum, body)
}
