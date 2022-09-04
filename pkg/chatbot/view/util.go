package view

import (
	"encoding/json"
	"strings"

	"github.com/slack-go/slack"
)

func castFromMapToMsg(m map[string]interface{}) (slack.Msg, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return slack.Msg{}, err
	}
	var result slack.Msg
	if err := json.Unmarshal(b, &result); err != nil {
		return slack.Msg{}, err
	}
	return result, nil
}

func castFromStringToMsg(s string) (slack.Msg, error) {
	var result slack.Msg
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return slack.Msg{}, err
	}
	return result, nil
}

func replaceBackquote(s string) string {
	return strings.ReplaceAll(s, "<backquote>", "`")
}
