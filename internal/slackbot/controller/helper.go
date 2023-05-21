package controller

import (
	"github.com/slack-go/slack"
)

func getCallbackValueOnStaticSelect(i slack.InteractionCallback) string {
	return i.ActionCallback.BlockActions[0].SelectedOption.Value
}

func getCallbackValueOnButton(i slack.InteractionCallback) string {
	return i.ActionCallback.BlockActions[0].Value
}
