package view

import (
	"fmt"
	"sort"
	"strings"

	"github.com/slack-go/slack"
)

func ShowCommands(commands map[string]string) slack.Msg {
	result, _ := showCommands(commands)
	return result
}

func showCommands(commands map[string]string) (slack.Msg, error) {
	var cmds []struct {
		command string
		url     string
	}
	for k, v := range commands {
		cmds = append(cmds, struct {
			command string
			url     string
		}{k, v})
	}
	sort.SliceStable(cmds, func(i, j int) bool { return cmds[i].command < cmds[j].command })

	var msg []string
	for _, c := range cmds {
		if c.url != "" {
			msg = append(msg, fmt.Sprintf("• <%s|%s>", c.url, c.command))
		} else {
			msg = append(msg, fmt.Sprintf("• %s", c.command))
		}
	}
	return castFromStringToMsg(replaceBackquote(fmt.Sprintf(`
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"verbatim": true,
				"text": "以下のコマンドが存在します。\n<backquote><backquote><backquote>%s<backquote><backquote><backquote>"
			}
		}
	]
}
`, strings.Join(msg, `\n`),
	)))
}

func SomethingIsWrong(messageTs string) slack.Msg {
	result, _ := somethingIsWrong(messageTs)
	return result
}

func somethingIsWrong(messageTs string) (slack.Msg, error) {
	return castFromMapToMsg(
		map[string]interface{}{
			"attachments": []interface{}{
				map[string]interface{}{
					"color": colorCrimson,
					"blocks": []interface{}{
						map[string]interface{}{
							"type": "section",
							"text": map[string]interface{}{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*InternalServerError*\n"+
									"Please confirm to application log (messageTs: `%s`)", messageTs),
							},
						},
					},
				},
			},
		},
	)
}

func Canceled() slack.Msg {
	result, _ := canceled()
	return result
}

func canceled() (slack.Msg, error) {
	return castFromStringToMsg(fmt.Sprintf(`
{
	"attachments": [
		{
			"color": "%s",
			"blocks": [
				{
					"type": "section",
					"text": {
						"type": "plain_text",
						"text": "キャンセルされました"
					}
				}
			]
		}
	]
}
`, colorHhaki))
}
