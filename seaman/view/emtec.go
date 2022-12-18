package view

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"sigs.k8s.io/yaml"

	pb "github.com/cloudnativedaysjp/emtec-ecu/pkg/ws-proxy/schema"
	"github.com/cloudnativedaysjp/seaman/seaman/api"
)

func EmtecListTrack(track []*pb.Track) slack.Msg {
	result, _ := emtecListTrack(track)
	return result
}

func emtecListTrack(pbTracks []*pb.Track) (slack.Msg, error) {
	type trackView struct {
		TrackId   int32  `json:"trackId"`
		TrackName string `json:"trackName"`
		ObsHost   string `json:"obsHost"`
		Enabled   bool   `json:"enabled"`
	}
	var track []trackView
	for _, pbTrack := range pbTracks {
		track = append(track, trackView{pbTrack.TrackId,
			pbTrack.TrackName, pbTrack.ObsHost, pbTrack.Enabled})
	}

	data, err := yaml.Marshal(&track)
	if err != nil {
		return slack.Msg{}, err
	}
	return castFromStringToMsg(replaceBackquote(fmt.Sprintf(`
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "<bq><bq><bq>%s<bq><bq><bq>"
			}
		}
	]
}
`, strings.ReplaceAll(string(data), "\n", "\\n"))))
}

func EmtecDisabled(trackName string) slack.Msg {
	result, _ := emtecDisabled(trackName)
	return result
}

func emtecDisabled(trackName string) (slack.Msg, error) {
	return castFromStringToMsg(fmt.Sprintf(`
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "Track %s の自動切り替えを無効化しました"
			}
		}
	]
}
`, trackName))
}

func EmtecEnabled(trackName string) slack.Msg {
	result, _ := emtecEnabled(trackName)
	return result
}

func emtecEnabled(trackName string) (slack.Msg, error) {
	return castFromStringToMsg(fmt.Sprintf(`
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "Track %s の自動切り替えを有効化しました"
			}
		}
	]
}
`, trackName))
}

func EmtecMovedToNextScene(msg slack.Msg) (slack.Msg, error) {
	// update "Move to Next Scene" Button
	var buttonValue string
	bs := &msg.Blocks.BlockSet
	secBlock, ok := (*bs)[len(*bs)-1].(*slack.SectionBlock)
	if !ok {
		return slack.Msg{}, fmt.Errorf("msg.Blocks.BlockSet[-1] cannot be cast to *slack.SectionBlock")
	}
	buttonValue = secBlock.Accessory.ButtonElement.Value
	secBlock.Accessory.ButtonElement = &slack.ButtonBlockElement{
		Type:     slack.METButton,
		ActionID: api.ActIdCommon_NothingToDo,
		Text: &slack.TextBlockObject{
			Type: "plain_text",
			Text: ":white_check_mark: Switched",
		},
	}
	// add "Set nextTalk OnAir" Button
	accessory := &slack.Accessory{
		ButtonElement: &slack.ButtonBlockElement{
			Type:     slack.METButton,
			ActionID: api.ActIdEmtec_OnAirNext,
			Value:    buttonValue,
			Text: &slack.TextBlockObject{
				Type:  "plain_text",
				Text:  "Switching",
				Emoji: true,
			},
			Style: slack.StylePrimary,
			Confirm: &slack.ConfirmationBlockObject{
				Title: &slack.TextBlockObject{
					Type: "plain_text",
					Text: "Set nextTalk OnAir",
				},
				Text: &slack.TextBlockObject{
					Type: "plain_text",
					Text: "Are you sure?",
				},
				Confirm: &slack.TextBlockObject{
					Type: "plain_text",
					Text: "OK",
				},
				Deny: &slack.TextBlockObject{
					Type: "plain_text",
					Text: "Cancel",
				},
			},
		},
	}
	msg.Blocks.BlockSet = append(msg.Blocks.BlockSet, slack.NewDividerBlock())
	msg.Blocks.BlockSet = append(msg.Blocks.BlockSet, &slack.SectionBlock{
		Type: slack.MBTSection,
		Text: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Set nextTalk OnAir",
		},
		Accessory: accessory,
	})

	return msg, nil
}

func EmtecMadeNextTalkOnAir(msg slack.Msg) (slack.Msg, error) {
	// update "Set nextTalk OnAir" Button
	bs := &msg.Blocks.BlockSet
	secBlock, ok := (*bs)[len(*bs)-1].(*slack.SectionBlock)
	if !ok {
		return slack.Msg{}, fmt.Errorf("msg.Blocks.BlockSet[-1] cannot be cast to *slack.SectionBlock")
	}
	secBlock.Accessory.ButtonElement = &slack.ButtonBlockElement{
		Type:     slack.METButton,
		ActionID: api.ActIdCommon_NothingToDo,
		Text: &slack.TextBlockObject{
			Type: "plain_text",
			Text: ":white_check_mark: Switched",
		},
	}
	return msg, nil
}
