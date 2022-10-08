package view

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_showVersion(t *testing.T) {
	t.Parallel()
	t.Run("test", func(t *testing.T) {
		expectedStr := replaceBackquote(`
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "Version REPLACEMENT (Commit: REPLACEMENT)\nRepoUrl: https://github.com/cloudnativedaysjp/seaman"
			}
		}
	]
}
`)
		expected, err := castFromStringToMsg(expectedStr)
		if err != nil {
			t.Fatal(err)
		}
		got, err := showVersion()
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf(diff)
		}
	})
}
