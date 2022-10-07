package view

import (
	"fmt"

	"github.com/slack-go/slack"

	"github.com/cloudnativedaysjp/seaman/version"
)

func ShowVersion() slack.Msg {
	result, _ := showVersion()
	return result
}

func showVersion() (slack.Msg, error) {
	return castFromStringToMsg(replaceBackquote(fmt.Sprintf(`
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "%s"
			}
		}
	]
}
`, version.Information(),
	)))
}
