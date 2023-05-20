package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/cloudnativedaysjp/emtec-ecu/pkg/ws-proxy/schema"

	infra_cnd "github.com/cloudnativedaysjp/seaman/internal/infra/emtec-ecu"
	infra_slack "github.com/cloudnativedaysjp/seaman/internal/infra/slack"
	"github.com/cloudnativedaysjp/seaman/internal/log"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/api"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/view"
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

func (c *EmtecController) ListTrack(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	ev, err := getAppMentionEvent(evt)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to get AppMentionEvent: %v", err))
		return
	}
	channelId := ev.Channel
	messageTs := ev.TimeStamp

	// init logger & context
	logger := c.log.With("messageTs", messageTs)
	ctx := log.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize Slack client: %v", err))
		return
	}

	resp, err := c.cndTrackClient.ListTrack(ctx, &emptypb.Empty{})
	if err != nil {
		logger.Error(fmt.Sprintf("cndTrackClient.ListTrack() was failed: %v", err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.PostMessage(ctx, channelId, view.EmtecListTrack(resp.Tracks)); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *EmtecController) EnableAutomation(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	c.switchAutomation(evt, client, true)
}

func (c *EmtecController) DisableAutomation(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)
	c.switchAutomation(evt, client, false)
}

func (c *EmtecController) switchAutomation(evt *socketmode.Event, client *socketmode.Client, enabled bool) {
	ev, err := getAppMentionEvent(evt)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to get AppMentionEvent: %v", err))
		return
	}
	channelId := ev.Channel
	messageTs := ev.TimeStamp

	// init logger & context
	logger := c.log.With("messageTs", messageTs)
	ctx := log.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize Slack client: %v", err))
		return
	}
	// parse arguments
	s := strings.Fields(ev.Text)
	if !(len(s) >= 4) {
		_ = sc.PostMessage(ctx, channelId, view.InvalidArguments(messageTs,
			"args.length must be greater than 2"))
		return
	}
	trackIdStr := s[3]
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
			logger.Error(fmt.Sprintf("cndTrackClient.DisableAutomation() was failed: %v", err))
			_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
			return
		}
		msg = view.EmtecEnabled(resp.TrackName)
	} else {
		resp, err := c.cndTrackClient.DisableAutomation(ctx,
			&pb.SwitchAutomationRequest{TrackId: int32(trackId)})
		if err != nil {
			logger.Error(fmt.Sprintf("cndTrackClient.DisableAutomation() was failed: %v", err))
			_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
			return
		}
		msg = view.EmtecDisabled(resp.TrackName)
	}

	if err := sc.PostMessage(ctx, channelId, msg); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}
}

func (c *EmtecController) UpdateSceneToNext(evt *socketmode.Event, client *socketmode.Client) {
	client.Ack(*evt.Request)

	interaction, err := getInteractionCallback(evt)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to get InteractionCallback: %v", err))
		return
	}
	channelId := interaction.Container.ChannelID
	messageTs := interaction.Container.MessageTs
	sentUserId := interaction.User.ID
	callbackValue := getCallbackValueOnButton(interaction)

	// init logger & context
	logger := c.log.With("messageTs", messageTs)
	ctx := log.IntoContext(context.Background(), logger)
	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize Slack client: %v", err))
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
		logger.Error(fmt.Sprintf("cndSceneClient.MoveScneToNext() was failed: %v", err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}

	msg, err := view.EmtecMovedToNextScene(interaction.Message.Msg)
	if err != nil {
		msg := "invalid interactive message"
		logger.Info(fmt.Sprintf("%s: %v", msg, err))
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.UpdateMessage(
		ctx, channelId, messageTs, msg,
	); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}

	if err := sc.PostMessageToThread(
		ctx, channelId, messageTs, slack.Msg{
			Text: fmt.Sprintf("Switching was pushed by <@%s>", sentUserId)},
	); err != nil {
		logger.Error(fmt.Sprintf("failed to post message: %v", err))
		_ = sc.UpdateMessage(ctx, channelId, messageTs, view.SomethingIsWrong(messageTs))
		return
	}
}
