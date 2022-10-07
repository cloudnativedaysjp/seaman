//go:generate go run github.com/golang/mock/mockgen -package mock -source=slack.go -destination=mock/slack.go

package slack

import (
	"context"

	"github.com/slack-go/slack"
	"golang.org/x/xerrors"
)

type SlackClient interface {
	PostMessage(ctx context.Context, channel string, msg slack.Msg) error
	UpdateMessage(ctx context.Context, channel, ts string, msg slack.Msg) error
}

type SlackClientImpl struct {
	client    slack.Client
	botUserId string
}

func NewSlackClientImpl(client slack.Client) (*SlackClientImpl, error) {
	res, err := client.AuthTest()
	if err != nil {
		return nil, err
	}

	return &SlackClientImpl{client, res.UserID}, nil
}

func (s *SlackClientImpl) PostMessage(ctx context.Context, channel string, msg slack.Msg) error {
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

func (s *SlackClientImpl) UpdateMessage(ctx context.Context, channel, ts string, msg slack.Msg) error {
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
