package slack

import "github.com/slack-go/slack"

type SlackDriverFactoryIface interface {
	New(client slack.Client) (SlackIface, error)
}

type SlackDriverFactory struct{}

func NewSlackDriverFactory() *SlackDriverFactory {
	return &SlackDriverFactory{}
}

func (f SlackDriverFactory) New(client slack.Client) (SlackIface, error) {
	return NewSlackDriver(client)
}
