package githubwh

import (
	"github.com/cloudnativedaysjp/seaman/internal/infra/gitcommand"
	"github.com/cloudnativedaysjp/seaman/internal/infra/githubapi"
	"github.com/cloudnativedaysjp/seaman/internal/service"
)

type Controller struct {
	gitcommand gitcommand.GitCommandClient
	githubapi  githubapi.GitHubApiClient
	service    service.GitHubIface
}

func NewController(
	gitcommand gitcommand.GitCommandClient,
	githubapi githubapi.GitHubApiClient,
) *Controller {
	service := service.NewGitHubService(gitcommand, githubapi)
	return &Controller{gitcommand, githubapi, service}
}
