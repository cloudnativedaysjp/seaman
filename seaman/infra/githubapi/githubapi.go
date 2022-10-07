//go:generate go run github.com/golang/mock/mockgen -package mock -source=githubapi.go -destination=mock/githubapi.go

package githubapi

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
)

type GitHubApiClient interface {
	HealthCheck() error
	CreatePullRequest(ctx context.Context, org, repo, headBranch, baseBranch, title, body string) (prNum int, err error)
	LabelPullRequest(ctx context.Context, org, repo string, prNum int, label string) error
	DeleteBranch(ctx context.Context, org, repo, headBranch string) error
}

type GitHubApiClientImpl struct {
	tokenSource oauth2.TokenSource
}

func NewGitHubApiClientImpl(token string) GitHubApiClient {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return &GitHubApiClientImpl{src}
}

func (g *GitHubApiClientImpl) HealthCheck() error {
	ctx := context.Background()
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))
	var q struct {
		Viewer struct {
			Login githubv4.String
		}
	}
	if err := client.Query(ctx, &q, nil); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	return nil
}

func (g *GitHubApiClientImpl) CreatePullRequest(ctx context.Context, org, repo, headBranch, baseBranch, title, body string) (prNum int, err error) {
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))

	repoId, err := g.getRepositoryId(ctx, org, repo)
	if err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}

	var mutationCreatePR struct {
		CreatePullRequest struct {
			PullRequest struct {
				Number int
			}
		} `graphql:"createPullRequest(input:$input)"`
	}
	if err := client.Mutate(ctx, &mutationCreatePR, githubv4.CreatePullRequestInput{
		RepositoryID: repoId,
		BaseRefName:  githubv4.String(baseBranch),
		HeadRefName:  githubv4.String(headBranch),
		Title:        githubv4.String(title),
		Body:         githubv4.NewString(githubv4.String(body)),
	}, nil); err != nil {
		return 0, xerrors.Errorf("message: %w", err)
	}

	return mutationCreatePR.CreatePullRequest.PullRequest.Number, nil
}

func (g *GitHubApiClientImpl) LabelPullRequest(ctx context.Context, org, repo string, prNum int, label string) error {
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))

	prId, err := g.getPullRequestId(ctx, org, repo, prNum)
	if err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	labelId, err := g.getLabelId(ctx, org, repo, label)
	if err != nil {
		return xerrors.Errorf("message: %w", err)
	}

	var mutationLabelPR struct {
		UpdatePullRequest struct {
			PullRequest struct {
				ResourcePath githubv4.URI
			}
		} `graphql:"updatePullRequest(input:$input)"`
	}
	if err := client.Mutate(ctx, &mutationLabelPR, githubv4.UpdatePullRequestInput{
		PullRequestID: prId,
		LabelIDs:      &[]githubv4.ID{labelId},
	}, nil); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	return nil
}

func (g *GitHubApiClientImpl) DeleteBranch(ctx context.Context, org, repo, headBranch string) error {
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))

	id, err := g.getBranchId(ctx, org, repo, headBranch)
	if err != nil {
		return xerrors.Errorf("message: %w", err)
	}

	var mutationDeleteBranch struct {
		DeleteRef struct {
			ClientMutationId githubv4.String
		} `graphql:"deleteRef(input:$input)"`
	}
	if err := client.Mutate(ctx, &mutationDeleteBranch, githubv4.DeleteRefInput{
		RefID: id,
	}, nil); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	return nil
}

func (g *GitHubApiClientImpl) getRepositoryId(ctx context.Context, org, repo string) (githubv4.ID, error) {
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))
	var queryGetRepository struct {
		Repository struct {
			ID githubv4.String
		} `graphql:"repository(owner:$repositoryOwner,name:$repositoryName)"`
	}
	if err := client.Query(ctx, &queryGetRepository, map[string]interface{}{
		"repositoryOwner": githubv4.String(org),
		"repositoryName":  githubv4.String(repo),
	}); err != nil {
		return 0, err
	}
	return queryGetRepository.Repository.ID, nil
}

func (g *GitHubApiClientImpl) getPullRequestId(ctx context.Context, org, repo string, prNum int) (githubv4.ID, error) {
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))
	var queryGetPullRequest struct {
		Repository struct {
			PullRequest struct {
				ID githubv4.ID
			} `graphql:"pullRequest(number:$pullRequestNumber)"`
		} `graphql:"repository(owner:$repositoryOwner,name:$repositoryName)"`
	}
	if err := client.Query(ctx, &queryGetPullRequest, map[string]interface{}{
		"repositoryOwner":   githubv4.String(org),
		"repositoryName":    githubv4.String(repo),
		"pullRequestNumber": githubv4.Int(prNum),
	}); err != nil {
		return 0, err
	}
	return queryGetPullRequest.Repository.PullRequest.ID, nil
}

func (g *GitHubApiClientImpl) getBranchId(ctx context.Context, org, repo, branch string) (githubv4.ID, error) {
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))
	var queryGetBranchID struct {
		Repository struct {
			Ref struct {
				ID githubv4.ID
			} `graphql:"ref(qualifiedName:$qualifiedName)"`
		} `graphql:"repository(owner:$repositoryOwner,name:$repositoryName)"`
	}
	if err := client.Query(ctx, &queryGetBranchID, map[string]interface{}{
		"repositoryOwner": githubv4.String(org),
		"repositoryName":  githubv4.String(repo),
		"qualifiedName":   githubv4.String(branch),
	}); err != nil {
		return nil, err
	}
	return queryGetBranchID.Repository.Ref.ID, nil
}

func (g *GitHubApiClientImpl) getLabelId(ctx context.Context, org, repo, label string) (githubv4.ID, error) {
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))
	var queryGetLabel struct {
		Repository struct {
			Label struct {
				ID githubv4.ID
			} `graphql:"label(name:$labelName)"`
		} `graphql:"repository(owner:$repositoryOwner,name:$repositoryName)"`
	}
	if err := client.Query(ctx, &queryGetLabel, map[string]interface{}{
		"repositoryOwner": githubv4.String(org),
		"repositoryName":  githubv4.String(repo),
		"labelName":       githubv4.String(label),
	}); err != nil {
		return nil, err
	}
	if queryGetLabel.Repository.Label.ID == nil {
		return nil, fmt.Errorf("no such label: %v", label)
	}
	return queryGetLabel.Repository.Label.ID, nil
}
