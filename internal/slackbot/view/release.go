package view

import (
	"fmt"
	"path/filepath"

	"github.com/slack-go/slack"

	"github.com/cloudnativedaysjp/seaman/internal/slackbot/api"
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
		map[string]any{
			"attachments": []any{
				map[string]any{
					"color": colorLightGray,
					"blocks": []any{
						map[string]any{
							"type": "section",
							"text": map[string]any{
								"type": "plain_text",
								"text": "リリース対象のリポジトリを選択",
							},
						},
						map[string]any{
							"type": "actions",
							"elements": []any{
								map[string]any{
									"type": "static_select",
									"placeholder": map[string]any{
										"type": "plain_text",
										"text": "Select an item",
									},
									"action_id": api.ActIdRelease_SelectedRepository,
									"options":   options,
								},
								map[string]any{
									"type": "button",
									"text": map[string]any{
										"type": "plain_text",
										"text": "Cancel",
									},
									"action_id": api.ActIdCommon_Cancel,
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

func ReleaseListLevel(orgRepo api.OrgRepo) slack.Msg {
	result, _ := releaseListLevel(orgRepo)
	return result
}

func releaseListLevel(orgRepo api.OrgRepo) (slack.Msg, error) {
	return castFromMapToMsg(
		map[string]any{
			"attachments": []any{
				map[string]any{
					"color": colorLightGray,
					"blocks": []any{
						map[string]any{
							"type": "section",
							"text": map[string]any{
								"text": "更新レベルを選択",
								"type": "plain_text",
							},
						},
						map[string]any{
							"type": "actions",
							"elements": []any{
								map[string]any{
									"type": "button",
									"text": map[string]any{
										"type": "plain_text",
										"text": api.CallbackValueRelease_VersionMajor,
									},
									"action_id": api.ActIdRelease_SelectedLevelMajor,
									"value":     orgRepo.WithLevel(api.CallbackValueRelease_VersionMajor).String(),
								},
								map[string]any{
									"type": "button",
									"text": map[string]any{
										"text": api.CallbackValueRelease_VersionMinor,
										"type": "plain_text",
									},
									"action_id": api.ActIdRelease_SelectedLevelMinor,
									"value":     orgRepo.WithLevel(api.CallbackValueRelease_VersionMinor).String(),
								},
								map[string]any{
									"type": "button",
									"text": map[string]any{
										"type": "plain_text",
										"text": api.CallbackValueRelease_VersionPatch,
									},
									"action_id": api.ActIdRelease_SelectedLevelPatch,
									"value":     orgRepo.WithLevel(api.CallbackValueRelease_VersionPatch).String(),
								},
								map[string]any{
									"type": "button",
									"text": map[string]any{
										"type": "plain_text",
										"text": "Cancel",
									},
									"action_id": api.ActIdCommon_Cancel,
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

func ReleaseConfirmation(orgRepoLevel api.OrgRepoLevel) slack.Msg {
	result, _ := releaseConfirmation(orgRepoLevel)
	return result
}

func releaseConfirmation(orgRepoLevel api.OrgRepoLevel) (slack.Msg, error) {
	org := orgRepoLevel.Org()
	repo := orgRepoLevel.Repo()
	level := orgRepoLevel.Level()
	return castFromMapToMsg(map[string]any{
		"attachments": []any{
			map[string]any{
				"color": colorLightGray,
				"blocks": []any{
					map[string]any{
						"type": "section",
						"text": map[string]any{
							"type": "mrkdwn",
							"text": fmt.Sprintf(
								"OK? > Target: *%s/%s*, Update Level: *%s*", org, repo, level,
							),
						},
					},
					map[string]any{
						"type": "actions",
						"elements": []any{
							map[string]any{
								"type": "button",
								"text": map[string]any{
									"type": "plain_text",
									"text": "OK",
								},
								"action_id": api.ActIdRelease_OK,
								"value":     orgRepoLevel.String(),
							},
							map[string]any{
								"type": "button",
								"text": map[string]any{
									"type": "plain_text",
									"text": "Cancel",
								},
								"action_id": api.ActIdCommon_Cancel,
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

func ReleaseDisplayPrLink(orgRepoLevel api.OrgRepoLevel, prNumber int) slack.Msg {
	result, _ := releaseDisplayPrLink(orgRepoLevel, prNumber)
	return result
}

func releaseDisplayPrLink(orgRepoLevel api.OrgRepoLevel, prNumber int) (slack.Msg, error) {
	org := orgRepoLevel.Org()
	repo := orgRepoLevel.Repo()
	level := orgRepoLevel.Level()
	return castFromMapToMsg(map[string]any{
		"attachments": []any{
			map[string]any{
				"color": colorDeepSkyBlue,
				"blocks": []any{
					map[string]any{
						"type": "section",
						"fields": []any{
							map[string]any{
								"type": "mrkdwn",
								"text": fmt.Sprintf("Target: *%s/%s*", org, repo),
							},
							map[string]any{
								"type": "mrkdwn",
								"text": fmt.Sprintf("Update Level: *%s*", level),
							},
						},
					},
					map[string]any{
						"type": "divider",
					},
					map[string]any{
						"type": "section",
						"text": map[string]any{
							"type": "mrkdwn",
							"text": fmt.Sprintf(":github: <%s>", orgRepoLevel.PullRequestUrl(prNumber)),
						},
					},
				},
			},
		},
	})
}
