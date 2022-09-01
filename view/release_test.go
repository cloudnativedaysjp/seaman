package view

import (
	"testing"

	"github.com/cloudnativedaysjp/chatbot/model"
	"github.com/google/go-cmp/cmp"
)

func Test_releaseListRepo(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		expectedStr := `
{
	"attachments": [
		{
			"color": "#d3d3d3",
			"blocks": [
				{
					"type": "section",
					"text": {
						"type": "plain_text",
						"text": "リリース対象のリポジトリを選択"
					}
				},
				{
					"type": "actions",
					"elements": [
						{
							"type": "static_select",
							"placeholder": {
								"type": "plain_text",
								"text": "Select an item"
							},
							"action_id": "release_selected_repo",
							"options": [
								{
									"text": {
										"type": "plain_text",
										"text": "cloudnativedaysjp/dreamkast"
									},
									"value": "cloudnativedaysjp__dreamkast"
								},
								{
									"text": {
										"type": "plain_text",
										"text": "cloudnativedaysjp/dreamkast-ui"
									},
									"value": "cloudnativedaysjp__dreamkast-ui"
								}
							]
						},
						{
							"type": "button",
							"text": {
								"type": "plain_text",
								"text": "Cancel"
							},
							"action_id": "common_cancel",
							"style": "danger"
						}
					]
				}
			]
		}
	]
}
`
		expected, err := castFromStringToMsg(expectedStr)
		if err != nil {
			t.Fatal(err)
		}
		got, err := releaseListRepo([]string{
			"https://github.com/cloudnativedaysjp/dreamkast",
			"https://github.com/cloudnativedaysjp/dreamkast-ui",
		})
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf(diff)
		}
	})
}

func Test_releaseListLevel(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		expectedStr := `
{
	"attachments": [
		{
			"color": "#d3d3d3",
			"blocks": [
				{
					"type": "section",
					"text": {
						"type": "plain_text",
						"text": "更新レベルを選択"
					}
				},
				{
					"type": "actions",
					"elements": [
						{
							"type": "button",
							"text": {
								"type": "plain_text",
								"text": "release/major"
							},
							"action_id": "release_selected_level_major",
							"value": "cloudnativedaysjp__dreamkast__release/major"
						},
						{
							"type": "button",
							"text": {
								"type": "plain_text",
								"text": "release/minor"
							},
							"action_id": "release_selected_level_minor",
							"value": "cloudnativedaysjp__dreamkast__release/minor"
						},
						{
							"type": "button",
							"text": {
								"type": "plain_text",
								"text": "release/patch"
							},
							"action_id": "release_selected_level_patch",
							"value": "cloudnativedaysjp__dreamkast__release/patch"
						},
						{
							"type": "button",
							"text": {
								"type": "plain_text",
								"text": "Cancel"
							},
							"action_id": "common_cancel",
							"style": "danger"
						}
					]
				}
			]
		}
	]
}
`
		expected, err := castFromStringToMsg(expectedStr)
		if err != nil {
			t.Fatal(err)
		}
		orgRepo, _ := model.NewOrgRepo("cloudnativedaysjp__dreamkast")
		got, err := releaseListLevel(orgRepo)
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf(diff)
		}
	})
}

func Test_releaseConfirmation(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		expectedStr := `
{
	"attachments": [
		{
			"color": "#d3d3d3",
			"blocks": [
				{
					"type": "section",
					"text": {
						"type": "mrkdwn",
						"text": "OK? > Target: *cloudnativedaysjp/dreamkast*, Update Level: *release/major*"
					}
				},
				{
					"type": "actions",
					"elements": [
						{
							"type": "button",
							"text": {
								"type": "plain_text",
								"text": "OK"
							},
							"action_id": "release_ok",
							"value": "cloudnativedaysjp__dreamkast__release/major"
						},
						{
							"type": "button",
							"text": {
								"type": "plain_text",
								"text": "Cancel"
							},
							"action_id": "common_cancel",
							"style": "danger"
						}
					]
				}
			]
		}
	]
}
`
		expected, err := castFromStringToMsg(expectedStr)
		if err != nil {
			t.Fatal(err)
		}
		orgRepoLevel, _ := model.NewOrgRepoLevel("cloudnativedaysjp__dreamkast__release/major")
		got, err := releaseConfirmation(orgRepoLevel)
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf(diff)
		}
	})
}

func Test_releaseProcessing(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		expectedStr := `
{
	"attachments": [
		{
			"color": "#d3d3d3",
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
`
		expected, err := castFromStringToMsg(expectedStr)
		if err != nil {
			t.Fatal(err)
		}
		got, err := releaseProcessing()
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf(diff)
		}
	})
}

func Test_releaseDisplayPrLink(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		expectedStr := `
{
	"attachments": [
		{
			"color": "#00bfff",
			"blocks": [
				{
					"type": "section",
					"fields": [
						{
							"type": "mrkdwn",
							"text": "Target: *cloudnativedaysjp/dreamkast*"
						},
						{
							"type": "mrkdwn",
							"text": "Update Level: *release/patch*"
						}
					]
				},
				{
					"type": "divider"
				},
				{
					"type": "section",
					"text": {
						"type": "mrkdwn",
						"text": ":github: <https://github.com/cloudnativedaysjp/dreamkast/pull/1416>"
					}
				}
			]
		}
	]
}
`
		expected, err := castFromStringToMsg(expectedStr)
		if err != nil {
			t.Fatal(err)
		}
		orgRepoLevel, _ := model.NewOrgRepoLevel("cloudnativedaysjp__dreamkast__release/patch")
		got, err := releaseDisplayPrLink(orgRepoLevel, 1416)
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf(diff)
		}
	})
}
