package controller

import (
	"fmt"

	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func getAppMentionEvent(evt *socketmode.Event) (*slackevents.AppMentionEvent, error) {
	eventsApiEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		return nil, fmt.Errorf("evt.Data cannot be casted to slackevents.EventsAPIEvent")
	}
	appMentionEvent, ok := eventsApiEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)
	if !ok {
		return nil, fmt.Errorf("eventsApiEvent.InnerEvent.Data cannot be casted to slackevents.AppMentionEvent")
	}
	return appMentionEvent, nil
}
