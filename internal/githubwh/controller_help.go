package githubwh

import (
	"context"

	"github.com/go-playground/webhooks/v6/github"
)

func (c Controller) CommandHelp(ctx context.Context, payload github.IssueCommentPayload, args []string) error {
	body := `以下のコマンドが存在します。
* /HELP
* /SEPARETE
`
	return c.githubapi.CreateIssueComment(ctx, payload.Repository.Owner.Login, payload.Repository.Name, int(payload.Issue.Number), body)
}
