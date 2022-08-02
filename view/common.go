package view

import (
	"fmt"

	"github.com/slack-go/slack"
)

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
