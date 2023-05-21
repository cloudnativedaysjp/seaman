package slackbot

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/cloudnativedaysjp/emtec-ecu/pkg/ws-proxy/schema"

	"github.com/cloudnativedaysjp/seaman/cmd/seaman/config"
	cndoperationserver "github.com/cloudnativedaysjp/seaman/internal/infra/emtec-ecu"
	"github.com/cloudnativedaysjp/seaman/internal/infra/gitcommand"
	"github.com/cloudnativedaysjp/seaman/internal/infra/githubapi"
	infra_slack "github.com/cloudnativedaysjp/seaman/internal/infra/slack"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/api"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/controller"
	"github.com/cloudnativedaysjp/seaman/pkg/lacks"
	seamanlog "github.com/cloudnativedaysjp/seaman/pkg/log"
)

func Run(ctx context.Context, conf *config.Config) error {
	logger := seamanlog.FromContext(ctx)

	// setup Slack Bot
	var client *socketmode.Client
	if conf.Debug {
		client = socketmode.New(
			slack.New(
				conf.Slack.BotToken,
				slack.OptionAppLevelToken(conf.Slack.AppToken),
				slack.OptionDebug(true),
				slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
			),
			socketmode.OptionDebug(true),
			socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
		)
	} else {
		client = socketmode.New(
			slack.New(
				conf.Slack.BotToken,
				slack.OptionAppLevelToken(conf.Slack.AppToken),
			),
		)
	}

	r := lacks.NewRouter(logger, client)

	// setup some instances
	slackFactory := infra_slack.NewSlackClientFactory()
	githubApiClient := githubapi.NewGitHubApiClientImpl(conf.GitHub.AccessToken)
	gitCommandClient := gitcommand.NewGitCommandClientImpl(conf.GitHub.Username, conf.GitHub.AccessToken)
	var cndClient *cndoperationserver.CndWrapper
	if conf.Emtec.EndpointUrl != "" {
		func() {
			conn, err := grpc.Dial(conf.Emtec.EndpointUrl,
				grpc.WithTransportCredentials(insecure.NewCredentials()), // TODO (cloudnativedaysjp/emtec-ecu#7)
			)
			if err != nil {
				logger.Warn(fmt.Sprintf("cannot connect to EMTEC-ECU, skipped: %v", err))
				return
			}
			cndClient = cndoperationserver.NewCndWrapper(
				pb.NewSceneServiceClient(conn), pb.NewTrackServiceClient(conn),
			)
		}()
	}

	{ // release
		var targets []controller.Target
		for _, target := range conf.Release.Targets {
			targets = append(targets, controller.Target(target))
		}
		c := controller.NewReleaseController(logger,
			slackFactory, gitCommandClient, githubApiClient, targets)
		// socketmodeHandler.HandleEvents(
		// 	slackevents.AppMention, middleware.MiddlewareSet(c.SelectRepository,
		// 		middleware.RegisterCommand("release").
		// 			WithURL("https://github.com/cloudnativedaysjp/seaman/blob/main/docs/release.md"),
		// 	))
		r.HandleMentionedMessage(
			"release", c.SelectRepository).
			WithURL("https://github.com/cloudnativedaysjp/seaman/blob/main/docs/release.md")
		r.HandleInteractionBlockAction(
			api.ActIdRelease_SelectedRepository, c.SelectReleaseLevel)
		r.HandleInteractionBlockAction(
			api.ActIdRelease_SelectedLevelMajor, c.SelectConfirmation)
		r.HandleInteractionBlockAction(
			api.ActIdRelease_SelectedLevelMinor, c.SelectConfirmation)
		r.HandleInteractionBlockAction(
			api.ActIdRelease_SelectedLevelPatch, c.SelectConfirmation)
		r.HandleInteractionBlockAction(
			api.ActIdRelease_OK, c.CreatePullRequestForRelease)
	}
	if cndClient != nil { // emtec
		c := controller.NewEmtecController(logger, slackFactory, cndClient)
		r.HandleMentionedMessage(
			"emtec list-track", c.ListTrack).
			WithURL("https://github.com/cloudnativedaysjp/seaman/blob/main/docs/emtec.md")
		r.HandleMentionedMessage(
			"emtec enable-track", c.EnableAutomation).
			WithURL("https://github.com/cloudnativedaysjp/seaman/blob/main/docs/emtec.md")
		r.HandleMentionedMessage(
			"emtec disable-track", c.DisableAutomation).
			WithURL("https://github.com/cloudnativedaysjp/seaman/blob/main/docs/emtec.md")
		r.HandleInteractionBlockAction(
			api.ActIdEmtec_SceneNext, c.UpdateSceneToNext)
	}
	{ // common
		c := controller.NewCommonController(logger,
			slackFactory)
		r.HandleHelp(c.ShowCommands)
		r.HandleMentionedMessage("version", c.ShowVersion)
		r.HandleInteractionBlockAction(
			api.ActIdCommon_NothingToDo, c.InteractionNothingToDo)
		r.HandleInteractionBlockAction(
			api.ActIdCommon_Cancel, c.InteractionCancel)
	}

	if err := r.RunEventLoop(); err != nil {
		return err
	}
	return nil
}
