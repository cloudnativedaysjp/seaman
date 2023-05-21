package utils

import (
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/xerrors"
)

func GetAppMentionEvent(evt *socketmode.Event) (*slackevents.AppMentionEvent, error) {
	eventsApiEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		return nil, xerrors.Errorf("evt.Data cannot be casted to slackevents.EventsAPIEvent")
	}
	appMentionEvent, ok := eventsApiEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)
	if !ok {
		return nil, xerrors.Errorf("eventsApiEvent.InnerEvent.Data cannot be casted to slackevents.AppMentionEvent")
	}
	return appMentionEvent, nil
}

func GetInteractionCallback(evt *socketmode.Event) (slack.InteractionCallback, error) {
	interaction, ok := evt.Data.(slack.InteractionCallback)
	if !ok {
		return slack.InteractionCallback{},
			xerrors.Errorf("evt.Data cannot be casted to slack.InteractionCallback")
	}
	return interaction, nil
}

func GetCallbackValueOnStaticSelect(i slack.InteractionCallback) string {
	return i.ActionCallback.BlockActions[0].SelectedOption.Value
}

func GetCallbackValueOnButton(i slack.InteractionCallback) string {
	return i.ActionCallback.BlockActions[0].Value
}
