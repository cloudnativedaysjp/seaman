package view

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	pb "github.com/cloudnativedaysjp/emtec-ecu/pkg/ws-proxy/schema"
)

func Test_emtecListTrack(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		expectedStr := replaceBackquote(`
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "<bq><bq><bq>- enabled: true\n  obsHost: https://a.example.com\n  trackId: 101\n  trackName: A\n- enabled: false\n  obsHost: https://b.example.com\n  trackId: 102\n  trackName: B\n<bq><bq><bq>"
			}
		}
	]
}
`)
		expected, err := castFromStringToMsg(expectedStr)
		if err != nil {
			t.Fatal(err)
		}
		track := []*pb.Track{
			{
				TrackId:   101,
				TrackName: "A",
				ObsHost:   "https://a.example.com",
				Enabled:   true,
			},
			{
				TrackId:   102,
				TrackName: "B",
				ObsHost:   "https://b.example.com",
				Enabled:   false,
			},
		}
		got, err := emtecListTrack(track)
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Error(diff)
		}
	})
}

func Test_EmtecMovedToNextScene(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		inputStr := `
{
	"blocks": [
		{
			"type": "divider"
		},
		{
			"type": "context",
			"elements": [
				{
					"type": "plain_text",
					"text": "Current Talk",
					"emoji": true
				}
			]
		},
		{
			"type": "section",
			"fields": [
				{
					"type": "plain_text",
					"text": "Track A",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "10:00 - 11:00",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "Type: オンライン登壇",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "Speaker: kanata",
					"emoji": true
				}
			]
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "Title: <https://event.cloudnativedays.jp/cndt2101/talks/10001|ものすごい発表>"
			}
		},
		{
			"type": "divider"
		},
		{
			"type": "context",
			"elements": [
				{
					"type": "plain_text",
					"text": "Next Talk",
					"emoji": true
				}
			]
		},
		{
			"type": "section",
			"fields": [
				{
					"type": "plain_text",
					"text": "Track A",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "11:00 - 12:30",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "Type: 事前収録",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "Speaker: hoge, fuga",
					"emoji": true
				}
			]
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "Title: <https://event.cloudnativedays.jp/cndt2101/talks/10002|さらにものすごい発表>"
			},
			"accessory": {
				"type": "button",
				"action_id": "emtec_scenenext",
				"value": "1__A",
				"text": {
					"type": "plain_text",
					"text": "Switching"
				},
				"style": "primary",
				"confirm": {
					"title": {
						"type": "plain_text",
						"text": "Move to Next Scene"
					},
					"text": {
						"type": "plain_text",
						"text": "Are you sure?"
					},
					"confirm": {
						"type": "plain_text",
						"text": "OK"
					},
					"deny": {
						"type": "plain_text",
						"text": "Cancel"
					}
				}
			}
		}
	]
}
`

		expectedStr := `
{
	"blocks": [
		{
			"type": "divider"
		},
		{
			"type": "context",
			"elements": [
				{
					"type": "plain_text",
					"text": "Current Talk",
					"emoji": true
				}
			]
		},
		{
			"type": "section",
			"fields": [
				{
					"type": "plain_text",
					"text": "Track A",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "10:00 - 11:00",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "Type: オンライン登壇",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "Speaker: kanata",
					"emoji": true
				}
			]
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "Title: <https://event.cloudnativedays.jp/cndt2101/talks/10001|ものすごい発表>"
			}
		},
		{
			"type": "divider"
		},
		{
			"type": "context",
			"elements": [
				{
					"type": "plain_text",
					"text": "Next Talk",
					"emoji": true
				}
			]
		},
		{
			"type": "section",
			"fields": [
				{
					"type": "plain_text",
					"text": "Track A",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "11:00 - 12:30",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "Type: 事前収録",
					"emoji": true
				},
				{
					"type": "plain_text",
					"text": "Speaker: hoge, fuga",
					"emoji": true
				}
			]
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "Title: <https://event.cloudnativedays.jp/cndt2101/talks/10002|さらにものすごい発表>"
			},
			"accessory": {
				"type": "button",
				"action_id": "common_nothing",
				"text": {
					"type": "plain_text",
					"text": ":white_check_mark: Switched"
				}
			}
		}
	]
}
`
		input, err := castFromStringToMsg(inputStr)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := castFromStringToMsg(expectedStr)
		if err != nil {
			t.Fatal(err)
		}
		got, err := EmtecMovedToNextScene(input)
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Error(diff)
		}
	})
}
