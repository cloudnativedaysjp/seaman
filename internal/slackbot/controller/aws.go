package controller

import (
	"context"
	"fmt"

	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/exp/slog"
	"golang.org/x/xerrors"

	"github.com/cloudnativedaysjp/seaman/internal/infra/aws"
	infra_slack "github.com/cloudnativedaysjp/seaman/internal/infra/slack"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot/view"
	"github.com/cloudnativedaysjp/seaman/pkg/log"
)

type AwsController struct {
	slackFactory infra_slack.SlackClientFactory
	awsClient    aws.AwsClient
	log          *slog.Logger
	rdsInstance  string
	ecsCluster   string
}

func NewAwsController(
	logger *slog.Logger,
	slackFactory infra_slack.SlackClientFactory,
	awsClient aws.AwsClient,
	rdsInstance string,
	ecsCluster string,
) *AwsController {
	return &AwsController{
		slackFactory: slackFactory,
		awsClient:    awsClient,
		log:          logger,
		rdsInstance:  rdsInstance,
		ecsCluster:   ecsCluster,
	}
}

func (c *AwsController) StartStaging(ctx context.Context, ev *slackevents.AppMentionEvent, client *socketmode.Client) error {
	logger := log.FromContext(ctx)
	channelId := ev.Channel
	messageTs := ev.TimeStamp

	// new client from factory
	sc, err := c.slackFactory.New(client.Client)
	if err != nil {
		return xerrors.Errorf("failed to initialize Slack client: %w", err)
	}

	// Post initial message
	if err := sc.PostMessage(ctx, channelId, view.AwsStartingStaging(messageTs)); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}

	// Start RDS instance
	logger.Info(fmt.Sprintf("Starting RDS instance: %s", c.rdsInstance))
	if err := c.awsClient.StartRdsInstance(ctx, c.rdsInstance); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.AwsStartRdsFailed(messageTs, err.Error()))
		return xerrors.Errorf("failed to start RDS instance: %w", err)
	}

	// Post RDS starting message
	if err := sc.PostMessage(ctx, channelId, view.AwsRdsStarting(messageTs, c.rdsInstance)); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}

	// Wait for RDS instance to be available
	logger.Info(fmt.Sprintf("Waiting for RDS instance to be available: %s", c.rdsInstance))
	if err := c.awsClient.WaitForRdsInstanceRunning(ctx, c.rdsInstance); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.AwsWaitRdsFailed(messageTs, err.Error()))
		return xerrors.Errorf("failed to wait for RDS instance: %w", err)
	}

	// Post RDS available message
	if err := sc.PostMessage(ctx, channelId, view.AwsRdsAvailable(messageTs, c.rdsInstance)); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}

	// Start all ECS services in the cluster
	logger.Info(fmt.Sprintf("Starting all ECS services in cluster %s", c.ecsCluster))
	services, err := c.awsClient.UpdateAllEcsServices(ctx, c.ecsCluster, 1)
	if err != nil {
		_ = sc.PostMessage(ctx, channelId, view.AwsStartEcsFailed(messageTs, err.Error()))
		return xerrors.Errorf("failed to start ECS services: %w", err)
	}

	// Post completion message
	if err := sc.PostMessage(ctx, channelId, view.AwsStartingComplete(messageTs, c.rdsInstance, c.ecsCluster, services)); err != nil {
		_ = sc.PostMessage(ctx, channelId, view.SomethingIsWrong(messageTs))
		return xerrors.Errorf("failed to post message: %w", err)
	}

	return nil
}
