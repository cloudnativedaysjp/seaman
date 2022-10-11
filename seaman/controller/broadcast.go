package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/xerrors"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/cloudnativedaysjp/cnd-operation-server/pkg/ws-proxy/schema"

	"github.com/cloudnativedaysjp/seaman/seaman/api"
	infra_cnd "github.com/cloudnativedaysjp/seaman/seaman/infra/cnd-operation-server"
	infra_slack "github.com/cloudnativedaysjp/seaman/seaman/infra/slack"
	"github.com/cloudnativedaysjp/seaman/seaman/utils"
	"github.com/cloudnativedaysjp/seaman/seaman/view"
)

type BroadcastController struct {
	slackFactory   infra_slack.SlackClientFactory
	cndSceneClient pb.SceneServiceClient
	cndTrackClient pb.TrackServiceClient
	log            logr.Logger
}

func NewBroadcastController(
	logger logr.Logger,
	slackFactory infra_slack.SlackClientFactory,
	cndClient *infra_cnd.CndWrapper,
) *BroadcastController {
	return &BroadcastController{slackFactory, cndClient, cndClient, logger}
}

func (c *BroadcastController) ListTrack(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	ev, err := getAppMentionEvent(evt)
	if err != nil {
		c.log.Error(err, "failed to get AppMentionEvent")
		return
	}
	channelId := ev.Channel
	messageTs := ev.TimeStamp

	// init logger & context
	logger := c.log.WithValues("messageTs", messageTs)
	ctx := utils.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err),
			"failed to initialize Slack client")
		return
	}

	resp, err := c.cndTrackClient.ListTrack(ctx, &emptypb.Empty{})
	if err != nil {
		logger.Error(err, "cndTrackClient.ListTrack() was failed")
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.PostMessage(ctx, channelId, view.BroadcastListTrack(resp.Tracks)); err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to post message")
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *BroadcastController) EnableAutomation(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	c.switchAutomation(evt, client, true)
}

func (c *BroadcastController) DisableAutomation(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	c.switchAutomation(evt, client, false)
}

func (c *BroadcastController) switchAutomation(evt *socketmode.Event, client *socketmode.Client, enabled bool) {
	ev, err := getAppMentionEvent(evt)
	if err != nil {
		c.log.Error(err, "failed to get AppMentionEvent")
		return
	}
	channelId := ev.Channel
	messageTs := ev.TimeStamp

	// init logger & context
	logger := c.log.WithValues("messageTs", messageTs)
	ctx := utils.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err),
			"failed to initialize Slack client")
		return
	}
	// parse arguments
	s := strings.Fields(ev.Text)
	if !(len(s) >= 2) {
		_ = sc.PostMessage(ctx, channelId, view.InvalidArguments(messageTs,
			"args.length must be greater than 2"))
		return
	}
	trackIdStr := s[1]
	trackId, err := strconv.Atoi(trackIdStr)
	if err != nil {
		_ = sc.PostMessage(ctx, channelId, view.InvalidArguments(messageTs,
			"args[1] (trackId) must be integer"))
		return
	}

	var msg slack.Msg
	if enabled {
		resp, err := c.cndTrackClient.EnableAutomation(ctx,
			&pb.SwitchAutomationRequest{TrackId: int32(trackId)})
		if err != nil {
			logger.Error(err, "cndTrackClient.DisableAutomation() was failed")
			_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
			return
		}
		msg = view.BroadcastEnabled(resp.TrackName)
	} else {
		resp, err := c.cndTrackClient.DisableAutomation(ctx,
			&pb.SwitchAutomationRequest{TrackId: int32(trackId)})
		if err != nil {
			logger.Error(err, "cndTrackClient.DisableAutomation() was failed")
			_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
			return
		}
		msg = view.BroadcastDisabled(resp.TrackName)
	}

	if err := sc.PostMessage(ctx, channelId, msg); err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to post message")
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *BroadcastController) UpdateSceneToNext(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)

	interaction, err := getInteractionCallback(evt)
	if err != nil {
		c.log.Error(err, "failed to get InteractionCallback")
		return
	}
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	callbackValue := getCallbackValueOnButton(interaction)

	// init logger & context
	logger := c.log.WithValues("messageTs", messageTs)
	ctx := utils.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to initialize Slack client")
	}

	track, err := api.NewTrack(callbackValue)
	if err != nil {
		_ = sc.PostMessage(ctx, channelId, view.InvalidArguments(messageTs,
			fmt.Sprintf("invalid format on callbackValue: %s", callbackValue)))
		return
	}

	if _, err := c.cndSceneClient.MoveSceneToNext(
		ctx, &pb.MoveSceneToNextRequest{TrackId: track.Id},
	); err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "cndSceneClient.MoveScneToNext() was failed")
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}

	msg, err := view.BroadcastMovedToNextScene(interaction.Message.Msg)
	if err != nil {
		msg := "invalid interactive message"
		logger.Info(fmt.Sprintf("%s: %v", msg, err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.UpdateMessage(
		ctx, channelId, messageTs, msg,
	); err != nil {
		logger.Error(xerrors.Errorf("message: %w", err), "failed to post message")
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}
