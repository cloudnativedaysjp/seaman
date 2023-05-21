package githubwh

import (
	"context"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/cloudnativedaysjp/seaman/cmd/seaman/config"
	"github.com/cloudnativedaysjp/seaman/internal/infra/gitcommand"
	"github.com/cloudnativedaysjp/seaman/internal/infra/githubapi"
	"github.com/cloudnativedaysjp/seaman/pkg/cosme"
	"github.com/cloudnativedaysjp/seaman/pkg/log"
)

// Run is entrypoint for runnging server for GitHub Webhook
func Run(ctx context.Context, conf *config.Config) error {
	logger := log.FromContext(ctx)

	// initialize
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(log.NewLoggerForChi(logger))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := log.IntoContext(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	githubApiClient := githubapi.NewGitHubApiClientImpl(conf.GitHub.AccessToken)
	gitCommandClient := gitcommand.NewGitCommandClientImpl(conf.GitHub.Username, conf.GitHub.AccessToken)
	c := NewController(gitCommandClient, githubApiClient)

	// wrapper for GitHub Webhook Server
	h, err := cosme.New(logger, conf.GitHubWebhook.Secret)
	if err != nil {
		return err
	}

	// routing
	r.Mount("/webhook/github", h.
		WithCommand("/HELP", c.CommandHelp).
		WithCommand("/SEPARATE", c.CommandSeparate))

	if err := http.ListenAndServe(conf.GitHubWebhook.BindAddr, r); err != nil {
		return err
	}
	return nil
}
