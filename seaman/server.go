package seaman

import (
	"log"
	"os"

	"github.com/go-logr/zapr"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/cloudnativedaysjp/cnd-operation-server/pkg/ws-proxy/schema"

	"github.com/cloudnativedaysjp/seaman/config"
	"github.com/cloudnativedaysjp/seaman/seaman/api"
	"github.com/cloudnativedaysjp/seaman/seaman/controller"
	cndoperationserver "github.com/cloudnativedaysjp/seaman/seaman/infra/cnd-operation-server"
	"github.com/cloudnativedaysjp/seaman/seaman/infra/gitcommand"
	"github.com/cloudnativedaysjp/seaman/seaman/infra/githubapi"
	infra_slack "github.com/cloudnativedaysjp/seaman/seaman/infra/slack"
	"github.com/cloudnativedaysjp/seaman/seaman/middleware"
)

func Run(conf *config.Config) error {
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
	socketmodeHandler := socketmode.NewSocketmodeHandler(client)

	// setup logger
	zapConf := zap.NewProductionConfig()
	zapConf.DisableStacktrace = true // due to output wrapped error in errorVerbose
	zapLogger, err := zapConf.Build()
	if err != nil {
		return err
	}
	logger := zapr.NewLogger(zapLogger)

	// setup some instances
	slackFactory := infra_slack.NewSlackClientFactory()
	githubApiClient := githubapi.NewGitHubApiClientImpl(conf.GitHub.AccessToken)
	gitCommandClient := gitcommand.NewGitCommandClientImpl(conf.GitHub.Username, conf.GitHub.AccessToken)
	var cndClient *cndoperationserver.CndWrapper
	if conf.Broadcast.EndpointUrl != "" {
		func() {
			conn, err := grpc.Dial(conf.Broadcast.EndpointUrl,
				grpc.WithTransportCredentials(insecure.NewCredentials()), // TODO (cloudnativedaysjp/cnd-operation-server#7)
			)
			if err != nil {
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
		socketmodeHandler.HandleEvents(
			slackevents.AppMention, middleware.MiddlewareSet(c.SelectRepository,
				middleware.RegisterCommand("release").
					WithURL("https://github.com/cloudnativedaysjp/seaman/blob/main/docs/release.md"),
			))
		socketmodeHandler.HandleInteractionBlockAction(
			api.ActIdRelease_SelectedRepository, c.SelectReleaseLevel)
		socketmodeHandler.HandleInteractionBlockAction(
			api.ActIdRelease_SelectedLevelMajor, c.SelectConfirmation)
		socketmodeHandler.HandleInteractionBlockAction(
			api.ActIdRelease_SelectedLevelMinor, c.SelectConfirmation)
		socketmodeHandler.HandleInteractionBlockAction(
			api.ActIdRelease_SelectedLevelPatch, c.SelectConfirmation)
		socketmodeHandler.HandleInteractionBlockAction(
			api.ActIdRelease_OK, c.CreatePullRequestForRelease)
	}
	if cndClient != nil { // broadcast
		c := controller.NewBroadcastController(logger, slackFactory, cndClient)
		socketmodeHandler.HandleEvents(
			slackevents.AppMention, middleware.MiddlewareSet(c.ListTrack,
				middleware.RegisterCommand("broadcast", "list-track").
					WithURL("https://github.com/cloudnativedaysjp/seaman/blob/main/docs/broadcast.md"),
			))
		socketmodeHandler.HandleEvents(
			slackevents.AppMention, middleware.MiddlewareSet(c.EnableAutomation,
				middleware.RegisterCommand("broadcast", "enable-track").
					WithURL("https://github.com/cloudnativedaysjp/seaman/blob/main/docs/broadcast.md"),
			))
		socketmodeHandler.HandleEvents(
			slackevents.AppMention, middleware.MiddlewareSet(c.DisableAutomation,
				middleware.RegisterCommand("broadcast", "disable-track").
					WithURL("https://github.com/cloudnativedaysjp/seaman/blob/main/docs/broadcast.md"),
			))
		socketmodeHandler.HandleInteractionBlockAction(
			api.ActIdBroadcast_SceneNext, c.UpdateSceneToNext)
	}
	{ // common
		c := controller.NewCommonController(logger,
			slackFactory, middleware.Subcommands.List())
		socketmodeHandler.HandleEvents(
			slackevents.AppMention, middleware.MiddlewareSet(
				c.ShowCommands, middleware.RegisterCommand("help")))
		socketmodeHandler.HandleEvents(
			slackevents.AppMention, middleware.MiddlewareSet(
				c.ShowVersion, middleware.RegisterCommand("version")))
		socketmodeHandler.HandleInteractionBlockAction(
			api.ActIdCommon_Cancel, c.InteractionCancel)
	}

	if err := socketmodeHandler.RunEventLoop(); err != nil {
		return err
	}
	return nil
}
