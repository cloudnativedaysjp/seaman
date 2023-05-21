package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/exp/slog"
	"golang.org/x/xerrors"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/cloudnativedaysjp/emtec-ecu/pkg/ws-proxy/schema"

	infra_cnd "github.com/cloudnativedaysjp/seaman/internal/infra/emtec-ecu"
	infra_slack "github.com/cloudnativedaysjp/seaman/internal/infra/slack"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/api"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/view"
	"github.com/cloudnativedaysjp/seaman/pkg/log"
	"github.com/cloudnativedaysjp/seaman/pkg/utils"
)

type EmtecController struct {
	slackFactory   infra_slack.SlackClientFactory
	cndSceneClient pb.SceneServiceClient
	cndTrackClient pb.TrackServiceClient
	log            *slog.Logger
}

func NewEmtecController(
	logger *slog.Logger,
	slackFactory infra_slack.SlackClientFactory,
	cndClient *infra_cnd.CndWrapper,
) *EmtecController {
	return &EmtecController{slackFactory, cndClient, cndClient, logger}
}

func (c *EmtecController) ListTrack(ctx context.Context, ev *slackevents.AppMentionEvent, client *socketmode.Client) error {
	channelId := ev.Channel
	messageTs := ev.TimeStamp

	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		return xerrors.Errorf("failed to initialize Slack client: %w", err)
	}

	resp, err := c.cndTrackClient.ListTrack(ctx, &emptypb.Empty{})
	if err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("cndTrackClient.ListTrack failed: %w", err)
	}

	if err := sc.PostMessage(ctx, channelId, view.EmtecListTrack(resp.Tracks)); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}
	return nil
}

func (c *EmtecController) EnableAutomation(ctx context.Context, ev *slackevents.AppMentionEvent, client *socketmode.Client) error {
	if err := c.switchAutomation(ctx, ev, client, true); err != nil {
		return xerrors.Errorf("%w", err)
	}
	return nil
}

func (c *EmtecController) DisableAutomation(ctx context.Context, ev *slackevents.AppMentionEvent, client *socketmode.Client) error {
	if err := c.switchAutomation(ctx, ev, client, false); err != nil {
		return xerrors.Errorf("%w", err)
	}
	return nil
}

func (c *EmtecController) switchAutomation(ctx context.Context, ev *slackevents.AppMentionEvent, client *socketmode.Client, enabled bool) error {
	logger := log.FromContext(ctx)
	channelId := ev.Channel
	messageTs := ev.TimeStamp

	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		return xerrors.Errorf("failed to initialize Slack client: %w", err)
	}
	// parse arguments
	s := strings.Fields(ev.Text)
	if !(len(s) >= 4) {
		msg := "args.length must be greater than 2"
		logger.Debug(fmt.Sprintf("invalid input: %v", msg))
		_ = sc.PostMessage(ctx, channelId, view.InvalidArguments(messageTs, msg))
		return nil
	}
	trackIdStr := s[3]
	trackId, err := strconv.Atoi(trackIdStr)
	if err != nil {
		msg := "args[1] (trackId) must be integer"
		logger.Debug(fmt.Sprintf("invalid input: %v", msg))
		_ = sc.PostMessage(ctx, channelId, view.InvalidArguments(messageTs, msg))
		return nil
	}

	var msg slack.Msg
	if enabled {
		resp, err := c.cndTrackClient.EnableAutomation(ctx,
			&pb.SwitchAutomationRequest{TrackId: int32(trackId)})
		if err != nil {
			_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
			return xerrors.Errorf("cndTrackClient.DisableAutomation failed: %w", err)
		}
		msg = view.EmtecEnabled(resp.TrackName)
	} else {
		resp, err := c.cndTrackClient.DisableAutomation(ctx,
			&pb.SwitchAutomationRequest{TrackId: int32(trackId)})
		if err != nil {
			_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
			return xerrors.Errorf("cndTrackClient.DisableAutomation failed: %w", err)
		}
		msg = view.EmtecDisabled(resp.TrackName)
	}

	if err := sc.PostMessage(ctx, channelId, msg); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}
	return nil
}

func (c *EmtecController) UpdateSceneToNext(ctx context.Context, interaction slack.InteractionCallback, client *socketmode.Client) error {
	logger := log.FromContext(ctx)
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	sentUserId := interaction.User.ID

	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		return xerrors.Errorf("failed to initialize Slack client: %w", err)
	}

	track, err := api.NewTrack(utils.GetCallbackValueOnButton(interaction))
	if err != nil {
		logger.Debug(fmt.Sprintf("invalid callback value: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return nil
	}

	if _, err := c.cndSceneClient.MoveSceneToNext(
		ctx, &pb.MoveSceneToNextRequest{TrackId: track.Id},
	); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("cndSceneClient.MoveScneToNext failed: %w", err)
	}

	msg, err := view.EmtecMovedToNextScene(interaction.Message.Msg)
	if err != nil {
		msg := "invalid interactive message"
		logger.Debug(fmt.Sprintf("invalid input: %v", msg))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return nil
	}

	if err := sc.UpdateMessage(
		ctx, channelId, messageTs, msg,
	); err != nil {
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}

	if err := sc.PostMessageToThread(
		ctx, channelId, messageTs, slack.Msg{
			Text: fmt.Sprintf("Switching was pushed by <@%s>", sentUserId)},
	); err != nil {
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}
	return nil
}
