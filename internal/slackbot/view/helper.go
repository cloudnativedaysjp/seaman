package view

import (
	"encoding/json"
	"strings"

	"github.com/slack-go/slack"
)

const (
	// Color
	colorLightGray   = "#d3d3d3"
	colorCrimson     = "#dc143c"
	colorDeepSkyBlue = "#00bfff"
	colorHhaki       = "#f0e68c"
)

func castFromMapToMsg(m map[string]any) (slack.Msg, error) {
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
	result := s
	result = strings.ReplaceAll(result, "<backquote>", "`")
	result = strings.ReplaceAll(result, "<bq>", "`")
	return result
}

func replaceNewLine(s string) string {
	return strings.ReplaceAll(s, "\n", "\\n")
}
