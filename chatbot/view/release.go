package view

import (
	"fmt"
	"path/filepath"

	"github.com/slack-go/slack"

	"github.com/cloudnativedaysjp/chatbot/chatbot/dto"
)

func ReleaseListRepo(repoUrls []string) slack.Msg {
	result, _ := releaseListRepo(repoUrls)
	return result
}

func releaseListRepo(repoUrls []string) (slack.Msg, error) {
	var options []*slack.OptionBlockObject
	for _, repoUrl := range repoUrls {
		repo := filepath.Base(repoUrl)
		org := filepath.Base(filepath.Dir(repoUrl))
		options = append(options,
			&slack.OptionBlockObject{
				Text: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: fmt.Sprintf("%s/%s", org, repo),
				},
				Value: fmt.Sprintf("%s__%s", org, repo),
			},
		)
	}

	return castFromMapToMsg(
		map[string]interface{}{
			"attachments": []interface{}{
				map[string]interface{}{
					"color": colorLightGray,
					"blocks": []interface{}{
						map[string]interface{}{
							"type": "section",
							"text": map[string]interface{}{
								"type": "plain_text",
								"text": "リリース対象のリポジトリを選択",
							},
						},
						map[string]interface{}{
							"type": "actions",
							"elements": []interface{}{
								map[string]interface{}{
									"type": "static_select",
									"placeholder": map[string]interface{}{
										"type": "plain_text",
										"text": "Select an item",
									},
									"action_id": dto.ActIdRelease_SelectedRepository,
									"options":   options,
								},
								map[string]interface{}{
									"type": "button",
									"text": map[string]interface{}{
										"type": "plain_text",
										"text": "Cancel",
									},
									"action_id": dto.ActIdCommon_Cancel,
									"style":     "danger",
								},
							},
						},
					},
				},
			},
		},
	)
}

func ReleaseListLevel(orgRepo dto.OrgRepo) slack.Msg {
	result, _ := releaseListLevel(orgRepo)
	return result
}

func releaseListLevel(orgRepo dto.OrgRepo) (slack.Msg, error) {
	return castFromMapToMsg(
		map[string]interface{}{
			"attachments": []interface{}{
				map[string]interface{}{
					"color": colorLightGray,
					"blocks": []interface{}{
						map[string]interface{}{
							"type": "section",
							"text": map[string]interface{}{
								"text": "更新レベルを選択",
								"type": "plain_text",
							},
						},
						map[string]interface{}{
							"type": "actions",
							"elements": []interface{}{
								map[string]interface{}{
									"type": "button",
									"text": map[string]interface{}{
										"type": "plain_text",
										"text": dto.CallbackValueRelease_VersionMajor,
									},
									"action_id": dto.ActIdRelease_SelectedLevelMajor,
									"value":     orgRepo.WithLevel(dto.CallbackValueRelease_VersionMajor).String(),
								},
								map[string]interface{}{
									"type": "button",
									"text": map[string]interface{}{
										"text": dto.CallbackValueRelease_VersionMinor,
										"type": "plain_text",
									},
									"action_id": dto.ActIdRelease_SelectedLevelMinor,
									"value":     orgRepo.WithLevel(dto.CallbackValueRelease_VersionMinor).String(),
								},
								map[string]interface{}{
									"type": "button",
									"text": map[string]interface{}{
										"type": "plain_text",
										"text": dto.CallbackValueRelease_VersionPatch,
									},
									"action_id": dto.ActIdRelease_SelectedLevelPatch,
									"value":     orgRepo.WithLevel(dto.CallbackValueRelease_VersionPatch).String(),
								},
								map[string]interface{}{
									"type": "button",
									"text": map[string]interface{}{
										"type": "plain_text",
										"text": "Cancel",
									},
									"action_id": dto.ActIdCommon_Cancel,
									"style":     "danger",
								},
							},
						},
					},
				},
			},
		},
	)
}

func ReleaseConfirmation(orgRepoLevel dto.OrgRepoLevel) slack.Msg {
	result, _ := releaseConfirmation(orgRepoLevel)
	return result
}

func releaseConfirmation(orgRepoLevel dto.OrgRepoLevel) (slack.Msg, error) {
	org := orgRepoLevel.Org()
	repo := orgRepoLevel.Repo()
	level := orgRepoLevel.Level()
	return castFromMapToMsg(map[string]interface{}{
		"attachments": []interface{}{
			map[string]interface{}{
				"color": colorLightGray,
				"blocks": []interface{}{
					map[string]interface{}{
						"type": "section",
						"text": map[string]interface{}{
							"type": "mrkdwn",
							"text": fmt.Sprintf(
								"OK? > Target: *%s/%s*, Update Level: *%s*", org, repo, level,
							),
						},
					},
					map[string]interface{}{
						"type": "actions",
						"elements": []interface{}{
							map[string]interface{}{
								"type": "button",
								"text": map[string]interface{}{
									"type": "plain_text",
									"text": "OK",
								},
								"action_id": dto.ActIdRelease_OK,
								"value":     orgRepoLevel.String(),
							},
							map[string]interface{}{
								"type": "button",
								"text": map[string]interface{}{
									"type": "plain_text",
									"text": "Cancel",
								},
								"action_id": dto.ActIdCommon_Cancel,
								"style":     "danger",
							},
						},
					},
				},
			},
		},
	})
}

func ReleaseProcessing() slack.Msg {
	result, _ := releaseProcessing()
	return result
}

func releaseProcessing() (slack.Msg, error) {
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
						"text": "processing..."
					}
				}
			]
		}
	]
}
`, colorLightGray))
}

func ReleaseDisplayPrLink(orgRepoLevel dto.OrgRepoLevel, prNumber int) slack.Msg {
	result, _ := releaseDisplayPrLink(orgRepoLevel, prNumber)
	return result
}

func releaseDisplayPrLink(orgRepoLevel dto.OrgRepoLevel, prNumber int) (slack.Msg, error) {
	org := orgRepoLevel.Org()
	repo := orgRepoLevel.Repo()
	level := orgRepoLevel.Level()
	return castFromMapToMsg(map[string]interface{}{
		"attachments": []interface{}{
			map[string]interface{}{
				"color": colorDeepSkyBlue,
				"blocks": []interface{}{
					map[string]interface{}{
						"type": "section",
						"fields": []interface{}{
							map[string]interface{}{
								"type": "mrkdwn",
								"text": fmt.Sprintf("Target: *%s/%s*", org, repo),
							},
							map[string]interface{}{
								"type": "mrkdwn",
								"text": fmt.Sprintf("Update Level: *%s*", level),
							},
						},
					},
					map[string]interface{}{
						"type": "divider",
					},
					map[string]interface{}{
						"type": "section",
						"text": map[string]interface{}{
							"type": "mrkdwn",
							"text": fmt.Sprintf(":github: <%s>", orgRepoLevel.PullRequestUrl(prNumber)),
						},
					},
				},
			},
		},
	})
}
