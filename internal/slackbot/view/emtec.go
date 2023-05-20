package view

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"sigs.k8s.io/yaml"

	pb "github.com/cloudnativedaysjp/emtec-ecu/pkg/ws-proxy/schema"

	"github.com/cloudnativedaysjp/seaman/internal/slackbot/api"
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
