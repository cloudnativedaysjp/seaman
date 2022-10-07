package slack

import "github.com/slack-go/slack"

type SlackClientFactory interface {
	New(client slack.Client) (SlackClient, error)
}

type SlackClientFactoryImpl struct{}

func NewSlackClientFactory() SlackClientFactory {
	return &SlackClientFactoryImpl{}
}

func (f SlackClientFactoryImpl) New(client slack.Client) (SlackClient, error) {
	return NewSlackClientImpl(client)
}
