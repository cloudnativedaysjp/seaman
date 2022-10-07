package slack

import (
	"context"

	"github.com/slack-go/slack"
	"golang.org/x/xerrors"
)

type SlackIface interface {
	PostMessage(ctx context.Context, channel string, msg slack.Msg) error
	UpdateMessage(ctx context.Context, channel, ts string, msg slack.Msg) error
}

type SlackDriver struct {
	client    slack.Client
	botUserId string
}

func NewSlackDriver(client slack.Client) (*SlackDriver, error) {
	res, err := client.AuthTest()
	if err != nil {
		return nil, err
	}

	return &SlackDriver{client, res.UserID}, nil
}

func (s *SlackDriver) PostMessage(ctx context.Context, channel string, msg slack.Msg) error {
	_, _, err := s.client.PostMessageContext(ctx, channel,
		slack.MsgOptionText(msg.Text, false),
		slack.MsgOptionAttachments(msg.Attachments...),
		slack.MsgOptionBlocks(msg.Blocks.BlockSet...),
	)
	if err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	return nil
}

func (s *SlackDriver) UpdateMessage(ctx context.Context, channel, ts string, msg slack.Msg) error {
	_, _, _, err := s.client.UpdateMessageContext(
		ctx, channel, ts,
		slack.MsgOptionText(msg.Text, false),
		slack.MsgOptionAttachments(msg.Attachments...),
		slack.MsgOptionBlocks(msg.Blocks.BlockSet...),
	)
	if err != nil {
		return xerrors.Errorf("message: %w", err)
	}
	return nil
}
