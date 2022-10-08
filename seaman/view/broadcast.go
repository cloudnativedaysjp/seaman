package view

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"sigs.k8s.io/yaml"

	pb "github.com/cloudnativedaysjp/cnd-operation-server/pkg/ws-proxy/schema"
	"github.com/cloudnativedaysjp/seaman/seaman/api"
)

func BroadcastListTrack(track []*pb.Track) slack.Msg {
	result, _ := broadcastListTrack(track)
	return result
}

func broadcastListTrack(pbTracks []*pb.Track) (slack.Msg, error) {
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

func BroadcastDisabled(trackName string) slack.Msg {
	result, _ := broadcastDisabled(trackName)
	return result
}

func broadcastDisabled(trackName string) (slack.Msg, error) {
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

func BroadcastEnabled(trackName string) slack.Msg {
	result, _ := broadcastEnabled(trackName)
	return result
}

func broadcastEnabled(trackName string) (slack.Msg, error) {
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

func BroadcastMovedToNextScene(track api.Track) slack.Msg {
	result, _ := broadcastMovedToNextScene(track)
	return result
}

func broadcastMovedToNextScene(track api.Track) (slack.Msg, error) {
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
`, track.Name))
}
