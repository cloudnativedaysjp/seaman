package view

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_showCommands(t *testing.T) {
	t.Parallel()
	t.Run("test", func(t *testing.T) {
		expectedStr := replaceBackquote(`
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "以下のコマンドが存在します。\n<backquote><backquote><backquote>• fuga\n• <https://example.com|hoge><backquote><backquote><backquote>"
			}
		}
	]
}
`)
		expected, err := castFromStringToMsg(expectedStr)
		if err != nil {
			t.Fatal(err)
		}
		got, err := showCommands(map[string]string{"hoge": "https://example.com", "fuga": ""})
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Error(diff)
		}
	})
}

func Test_somethingIsWrong(t *testing.T) {
	t.Parallel()
	t.Run("test", func(t *testing.T) {
		expectedStr := replaceBackquote(`
{
	"attachments": [
		{
			"color": "#dc143c",
			"blocks": [
				{
					"type": "section",
					"text": {
						"type": "mrkdwn",
						"text": "*InternalServerError*\nPlease confirm to application log (messageTs: <backquote>12345678<backquote>)"
					}
				}
			]
		}
	]
}
`)
		expected, err := castFromStringToMsg(expectedStr)
		if err != nil {
			t.Fatal(err)
		}
		got, err := somethingIsWrong("12345678")
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Error(diff)
		}
	})
}

func Test_canceled(t *testing.T) {
	t.Parallel()
	t.Run("test", func(t *testing.T) {
		expectedStr := `
{
	"attachments": [
		{
			"color": "#f0e68c",
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
`
		expected, err := castFromStringToMsg(expectedStr)
		if err != nil {
			t.Fatal(err)
		}
		got, err := canceled()
		if err != nil {
			t.Errorf("error = %v", err)
			return
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Error(diff)
		}
	})
}
