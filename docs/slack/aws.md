# AWS Commands

This document describes the AWS-related commands available in the Seaman SlackBot.

## start-staging

The `start-staging` command starts the staging environment by:

1. Starting the RDS instance
2. Waiting for the RDS instance to be available
3. Starting all ECS services in the specified cluster

### Usage

```
@seaman start-staging
```

### Configuration

To use this command, you need to configure the following in your `config.yaml`:

```yaml
aws:
  rdsInstance: "your-rds-instance-identifier"
  ecsCluster: "your-ecs-cluster-name"
```

### AWS Permissions

The AWS credentials used by Seaman need the following permissions:

- `rds:StartDBInstance`
- `rds:DescribeDBInstances`
- `ecs:UpdateService`
- `ecs:DescribeServices`

### Example

```
User: @seaman start-staging
Seaman: 🚀 Starting staging environment
        I'll start the RDS instance and then the ECS service.

Seaman: ⏳ RDS instance starting
        Started RDS instance `staging-db`. Waiting for it to become available...

Seaman: ✅ RDS instance available
        RDS instance `staging-db` is now available. Starting all ECS services in the cluster...

Seaman: 🎉 Staging environment is ready!
        • RDS instance: `staging-db`
        • ECS cluster: `staging-cluster`
        • Started ECS services:
          • `service-1`
          • `service-2`
          • `service-3`
        
        The staging environment is now up and running.