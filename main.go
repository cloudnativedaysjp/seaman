package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/cloudnativedaysjp/slackbot/controller"
	"github.com/cloudnativedaysjp/slackbot/global"
	"github.com/cloudnativedaysjp/slackbot/infrastructure/gitcommand"
	"github.com/cloudnativedaysjp/slackbot/infrastructure/githubapi"
	slack_driver "github.com/cloudnativedaysjp/slackbot/infrastructure/slack"
)

func main() {
	var confFile string
	flag.StringVar(&confFile, "config", "", "")
	flag.Parse()
	if confFile == "" {
		fmt.Println("flag --config must be specified")
		os.Exit(1)
	}
	conf, err := loadConf(confFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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

	// setup some Drivers
	slackDriverFactory := slack_driver.NewSlackDriverFactory()
	githubApiDriver := githubapi.NewGitHubApiDriver(conf.GitHub.AccessToken)
	gitCommandDriver := gitcommand.NewGitCommandDriver(conf.GitHub.Username, conf.GitHub.AccessToken)

	{ // release
		c := controller.NewReleaseController(
			slackDriverFactory, gitCommandDriver, githubApiDriver,
			conf.Release.Targets, conf.Release.BaseBranch)
		socketmodeHandler.HandleEvents(
			slackevents.AppMention, middlewareMessagePrefixIs("release", c.SelectRepository))
		socketmodeHandler.HandleInteractionBlockAction(
			global.ActIdRelease_SelectedRepository, c.SelectReleaseLevel)
		socketmodeHandler.HandleInteractionBlockAction(
			global.ActIdRelease_SelectedLevelMajor, c.SelectConfirmation)
		socketmodeHandler.HandleInteractionBlockAction(
			global.ActIdRelease_SelectedLevelMinor, c.SelectConfirmation)
		socketmodeHandler.HandleInteractionBlockAction(
			global.ActIdRelease_SelectedLevelPatch, c.SelectConfirmation)
		socketmodeHandler.HandleInteractionBlockAction(
			global.ActIdRelease_OK, c.CreatePullRequestForRelease)
	}
	{ // common (THIS MUST BE DECLARED AT THE END)
		var cmds []string
		for cmd := range subcommands {
			cmds = append(cmds, cmd)
		}
		sort.SliceStable(cmds, func(i, j int) bool { return cmds[i] < cmds[j] })
		c := controller.NewCommonController(slackDriverFactory, cmds)
		socketmodeHandler.HandleEvents(
			slackevents.AppMention, middlewareMessagePrefixIs("help", c.ShowCommands))
		socketmodeHandler.HandleInteractionBlockAction(
			global.ActIdCommon_Cancel, c.InteractionCancel)
	}

	socketmodeHandler.RunEventLoop()
}
