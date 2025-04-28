package aws

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"golang.org/x/xerrors"
)

type AwsClient interface {
	StartRdsInstance(ctx context.Context, instanceIdentifier string) error
	WaitForRdsInstanceRunning(ctx context.Context, instanceIdentifier string) error
	UpdateEcsService(ctx context.Context, cluster string, service string, desiredCount int32) error
	ListEcsServices(ctx context.Context, cluster string) ([]string, error)
	UpdateAllEcsServices(ctx context.Context, cluster string, desiredCount int32) ([]string, error)
}

type AwsClientImpl struct {
	rdsClient *rds.Client
	ecsClient *ecs.Client
}

func NewAwsClient() (*AwsClientImpl, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, xerrors.Errorf("failed to load AWS config: %w", err)
	}

	rdsClient := rds.NewFromConfig(cfg)
	ecsClient := ecs.NewFromConfig(cfg)

	return &AwsClientImpl{
		rdsClient: rdsClient,
		ecsClient: ecsClient,
	}, nil
}

func (c *AwsClientImpl) StartRdsInstance(ctx context.Context, instanceIdentifier string) error {
	_, err := c.rdsClient.StartDBInstance(ctx, &rds.StartDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceIdentifier),
	})
	if err != nil {
		return xerrors.Errorf("failed to start RDS instance: %w", err)
	}
	return nil
}

func (c *AwsClientImpl) WaitForRdsInstanceRunning(ctx context.Context, instanceIdentifier string) error {
	waiter := rds.NewDBInstanceAvailableWaiter(c.rdsClient)
	err := waiter.Wait(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instanceIdentifier),
	}, 10*time.Minute)
	if err != nil {
		return xerrors.Errorf("failed to wait for RDS instance to be available: %w", err)
	}
	return nil
}

func (c *AwsClientImpl) UpdateEcsService(ctx context.Context, cluster string, service string, desiredCount int32) error {
	_, err := c.ecsClient.UpdateService(ctx, &ecs.UpdateServiceInput{
		Cluster:      aws.String(cluster),
		Service:      aws.String(service),
		DesiredCount: aws.Int32(desiredCount),
	})
	if err != nil {
		return xerrors.Errorf("failed to update ECS service: %w", err)
	}
	return nil
}

func (c *AwsClientImpl) ListEcsServices(ctx context.Context, cluster string) ([]string, error) {
	var serviceArns []string
	var nextToken *string

	for {
		output, err := c.ecsClient.ListServices(ctx, &ecs.ListServicesInput{
			Cluster:   aws.String(cluster),
			NextToken: nextToken,
		})
		if err != nil {
			return nil, xerrors.Errorf("failed to list ECS services: %w", err)
		}

		serviceArns = append(serviceArns, output.ServiceArns...)

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	// Extract service names from ARNs
	var serviceNames []string
	for _, arn := range serviceArns {
		// Extract the service name from the ARN
		// ARN format: arn:aws:ecs:region:account-id:service/cluster-name/service-name
		parts := strings.Split(arn, "/")
		if len(parts) > 0 {
			serviceNames = append(serviceNames, parts[len(parts)-1])
		}
	}

	return serviceNames, nil
}

func (c *AwsClientImpl) UpdateAllEcsServices(ctx context.Context, cluster string, desiredCount int32) ([]string, error) {
	services, err := c.ListEcsServices(ctx, cluster)
	if err != nil {
		return nil, xerrors.Errorf("failed to list ECS services: %w", err)
	}

	var updatedServices []string
	for _, service := range services {
		err := c.UpdateEcsService(ctx, cluster, service, desiredCount)
		if err != nil {
			return updatedServices, xerrors.Errorf("failed to update ECS service %s: %w", service, err)
		}
		updatedServices = append(updatedServices, service)
	}

	return updatedServices, nil
}