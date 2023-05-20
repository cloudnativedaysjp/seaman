//go:generate go run github.com/golang/mock/mockgen -package mock -source=githubapi.go -destination=mock/githubapi.go

package githubapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudnativedaysjp/seaman/internal/log"
	"github.com/cloudnativedaysjp/seaman/pkg/utils"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
)

type GitHubApiClient interface {
	CheckPrIsForInfraAndCreatedByRenovate(ctx context.Context, org, repo string, prNum int) (bool, string, error)
	CreateIssueComment(ctx context.Context, org, repo string, prNum int, body string) error
	CreateLabels(ctx context.Context, org, repo string, prNum int, labels []string) error
	CreatePullRequest(ctx context.Context, org, repo, headBranch, baseBranch, title, body string) (prNum int, err error)
	DeleteBranch(ctx context.Context, org, repo, headBranch string) error
	GetPullRequestTitleAndChangedFilepaths(ctx context.Context, org, repo string, prNum int) (string, []string, error)
	HealthCheck() error
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

//
// Exposed methods
//

func (g *GitHubApiClientImpl) CheckPrIsForInfraAndCreatedByRenovate(ctx context.Context, org, repo string, prNum int) (bool, string, error) {
	logger := log.FromContext(ctx)
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))
	expectedNumOfUpdatedFiles := 2
	labelLimit := 10

	var query struct {
		Repository struct {
			PullRequest struct {
				HeadRefName  githubv4.String
				ChangedFiles githubv4.Int
				Files        struct {
					Edges []struct {
						Node struct {
							Path githubv4.String
						}
					}
				} `graphql:"files(first:$filesFirst)"`
				Labels struct {
					Edges []struct {
						Node struct {
							Name githubv4.String
						}
					}
				} `graphql:"labels(first:$labelsFirst)"`
			} `graphql:"pullRequest(number:$number)"`
		} `graphql:"repository(owner:$repositoryOwner,name:$repositoryName)"`
	}
	queryVars := map[string]interface{}{
		"repositoryOwner": githubv4.String(org),
		"repositoryName":  githubv4.String(repo),
		"number":          githubv4.Int(prNum),
		"filesFirst":      githubv4.Int(expectedNumOfUpdatedFiles),
		"labelsFirst":     githubv4.Int(labelLimit),
	}
	if err := client.Query(ctx, &query, queryVars); err != nil {
		return false, "", err
	}
	// if changeFiles == `expectedNumOfUpdatedFiles`
	if int(query.Repository.PullRequest.ChangedFiles) != expectedNumOfUpdatedFiles {
		logger.Info(fmt.Sprintf("changeFiles != %d", expectedNumOfUpdatedFiles))
		return false, "", nil
	}
	// if path of changed file contains "/development/" or "/production/"
	fpath := string(query.Repository.PullRequest.Files.Edges[0].Node.Path)
	if !(strings.Contains(fpath, "/development/") || strings.Contains(fpath, "/production/")) {
		logger.Info(`path of changed file does contain neither "/development/" "/production/"`)
		return false, "", nil
	}
	// if pr labels contains "dependencies"
	actualLabels := []string{}
	for _, edge := range query.Repository.PullRequest.Labels.Edges {
		actualLabels = append(actualLabels, string(edge.Node.Name))
	}
	if !utils.Contains(actualLabels, "dependencies") {
		logger.Info(`PR is not labeled "dependencies" Label`)
		return false, "", nil
	}

	return true, string(query.Repository.PullRequest.HeadRefName), nil
}

func (g *GitHubApiClientImpl) CreateIssueComment(ctx context.Context, org, repo string, prNum int, body string) error {
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))

	prId, err := g.getPullRequestId(ctx, org, repo, prNum)
	if err != nil {
		return xerrors.Errorf("message: %w", err)
	}

	var mutationAddComment struct {
		AddComment struct {
			Subject struct {
				ID githubv4.ID
			}
		} `graphql:"addComment(input:$input)"`
	}
	if err := client.Mutate(ctx, &mutationAddComment, githubv4.AddCommentInput{
		SubjectID: prId,
		Body:      githubv4.String(body),
	}, nil); err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	return nil
}

func (g *GitHubApiClientImpl) CreateLabels(ctx context.Context, org, repo string, prNum int, labels []string) error {
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))

	prId, err := g.getPullRequestId(ctx, org, repo, prNum)
	if err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	labelIds := []githubv4.ID{}
	for _, label := range labels {
		labelId, err := g.getLabelId(ctx, org, repo, label)
		if err != nil {
			return xerrors.Errorf("message: %w", err)
		}
		labelIds = append(labelIds, labelId)
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
		LabelIDs:      &labelIds,
	}, nil); err != nil {
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

func (g *GitHubApiClientImpl) GetPullRequestTitleAndChangedFilepaths(ctx context.Context, org, repo string, prNum int) (string, []string, error) {
	client := githubv4.NewClient(oauth2.NewClient(ctx, g.tokenSource))
	pageLimit := 10

	var query struct {
		Repository struct {
			PullRequest struct {
				Title        githubv4.String
				ChangedFiles githubv4.Int
				Files        struct {
					Edges []struct {
						Cursor githubv4.String
						Node   struct {
							Path githubv4.String
						}
					}
				} `graphql:"files(first:$first,after:$after)"`
			} `graphql:"pullRequest(number:$pullRequestNumber)"`
		} `graphql:"repository(owner:$repositoryOwner,name:$repositoryName)"`
	}
	queryVars := map[string]interface{}{
		"repositoryOwner":   githubv4.String(org),
		"repositoryName":    githubv4.String(repo),
		"pullRequestNumber": githubv4.Int(prNum),
		"first":             githubv4.Int(pageLimit),
		"after":             githubv4.String(""),
	}

	title := ""
	changedFilesNum := 1
	changedFiles := []string{}
	for i := 0; i*pageLimit < changedFilesNum; i++ {
		if err := client.Query(ctx, &query, queryVars); err != nil {
			return "", nil, err
		}
		for _, edge := range query.Repository.PullRequest.Files.Edges {
			changedFiles = append(changedFiles, string(edge.Node.Path))
		}
		if len(query.Repository.PullRequest.Files.Edges) == pageLimit {
			queryVars["after"] = query.Repository.PullRequest.Files.Edges[pageLimit].Cursor
		}
		title = string(query.Repository.PullRequest.Title)
		changedFilesNum = int(query.Repository.PullRequest.ChangedFiles)
	}

	return title, changedFiles, nil
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

//
// Unexposed methods
//

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
