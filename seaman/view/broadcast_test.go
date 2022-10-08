package view

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	pb "github.com/cloudnativedaysjp/cnd-operation-server/pkg/ws-proxy/schema"
)

func Test_broadcastListTrack(t *testing.T) {
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
		got, err := broadcastListTrack(track)
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf(diff)
		}
	})
}
