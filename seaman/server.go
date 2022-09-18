package seaman

import (
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"github.com/cloudnativedaysjp/seaman/seaman/controller"
	"github.com/cloudnativedaysjp/seaman/seaman/dto"
	"github.com/cloudnativedaysjp/seaman/seaman/infrastructure/gitcommand"
	"github.com/cloudnativedaysjp/seaman/seaman/infrastructure/githubapi"
	infra_slack "github.com/cloudnativedaysjp/seaman/seaman/infrastructure/slack"
	"github.com/cloudnativedaysjp/seaman/seaman/middleware"
)

func Run(conf *Config) error {
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
	logger, err := zapConf.Build()
	if err != nil {
		return err
	}

	// setup some instances
	slackDriverFactory := infra_slack.NewSlackDriverFactory()
	githubApiDriver := githubapi.NewGitHubApiDriver(conf.GitHub.AccessToken)
	gitCommandDriver := gitcommand.NewGitCommandDriver(conf.GitHub.Username, conf.GitHub.AccessToken)

	{ // release
		var targets []controller.Target
		for _, target := range conf.Release.Targets {
			targets = append(targets, controller.Target(target))
		}
		c := controller.NewReleaseController(logger,
			slackDriverFactory, gitCommandDriver, githubApiDriver, targets)

		socketmodeHandler.HandleEvents(
			slackevents.AppMention, middleware.MiddlewareSet(
				c.SelectRepository,
				middleware.MiddlewareMessagePrefixIs{Prefix: "release"},
				middleware.MiddlewareHelpMessage{
					Prefix: "release",
					URL:    "https://github.com/cloudnativedaysjp/seaman/blob/main/docs/release.md",
				},
			))
		socketmodeHandler.HandleInteractionBlockAction(
			dto.ActIdRelease_SelectedRepository, c.SelectReleaseLevel)
		socketmodeHandler.HandleInteractionBlockAction(
			dto.ActIdRelease_SelectedLevelMajor, c.SelectConfirmation)
		socketmodeHandler.HandleInteractionBlockAction(
			dto.ActIdRelease_SelectedLevelMinor, c.SelectConfirmation)
		socketmodeHandler.HandleInteractionBlockAction(
			dto.ActIdRelease_SelectedLevelPatch, c.SelectConfirmation)
		socketmodeHandler.HandleInteractionBlockAction(
			dto.ActIdRelease_OK, c.CreatePullRequestForRelease)
	}
	{ // common (THIS MUST BE DECLARED AT THE END)
		c := controller.NewCommonController(logger,
			slackDriverFactory, middleware.Subcommands.List())

		socketmodeHandler.HandleEvents(
			slackevents.AppMention, middleware.MiddlewareSet(
				c.ShowCommands,
				middleware.MiddlewareMessagePrefixIs{Prefix: "help"},
			))
		socketmodeHandler.HandleInteractionBlockAction(
			dto.ActIdCommon_Cancel, c.InteractionCancel)
	}

	if err := socketmodeHandler.RunEventLoop(); err != nil {
		return err
	}
	return nil
}