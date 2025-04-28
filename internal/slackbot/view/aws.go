package view

import (
	"fmt"

	"github.com/slack-go/slack"
)

func AwsStartingStaging(messageTs string) slack.Msg {
	return slack.Msg{
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject(
						slack.MarkdownType,
						":rocket: *Starting staging environment*\n"+
							"I'll start the RDS instance and then the ECS service.",
						false, false),
					nil, nil,
				),
			},
		},
	}
}

func AwsRdsStarting(messageTs string, rdsInstance string) slack.Msg {
	return slack.Msg{
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject(
						slack.MarkdownType,
						fmt.Sprintf(":hourglass_flowing_sand: *RDS instance starting*\n"+
							"Started RDS instance `%s`. Waiting for it to become available...", rdsInstance),
						false, false),
					nil, nil,
				),
			},
		},
	}
}

func AwsRdsAvailable(messageTs string, rdsInstance string) slack.Msg {
	return slack.Msg{
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject(
						slack.MarkdownType,
						fmt.Sprintf(":white_check_mark: *RDS instance available*\n"+
							"RDS instance `%s` is now available. Starting all ECS services in the cluster...", rdsInstance),
						false, false),
					nil, nil,
				),
			},
		},
	}
}

func AwsStartingComplete(messageTs string, rdsInstance string, ecsCluster string, ecsServices []string) slack.Msg {
	// Format the list of services
	var servicesList string
	for i, service := range ecsServices {
		servicesList += fmt.Sprintf("  • `%s`", service)
		if i < len(ecsServices)-1 {
			servicesList += "\n"
		}
	}

	return slack.Msg{
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject(
						slack.MarkdownType,
						fmt.Sprintf(":tada: *Staging environment is ready!*\n"+
							"• RDS instance: `%s`\n"+
							"• ECS cluster: `%s`\n"+
							"• Started ECS services:\n%s\n\n"+
							"The staging environment is now up and running.",
							rdsInstance, ecsCluster, servicesList),
						false, false),
					nil, nil,
				),
			},
		},
	}
}

func AwsStartRdsFailed(messageTs string, errorMessage string) slack.Msg {
	return slack.Msg{
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject(
						slack.MarkdownType,
						fmt.Sprintf(":x: *Failed to start RDS instance*\n"+
							"Error: %s", errorMessage),
						false, false),
					nil, nil,
				),
			},
		},
	}
}

func AwsWaitRdsFailed(messageTs string, errorMessage string) slack.Msg {
	return slack.Msg{
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject(
						slack.MarkdownType,
						fmt.Sprintf(":x: *Failed while waiting for RDS instance to become available*\n"+
							"Error: %s", errorMessage),
						false, false),
					nil, nil,
				),
			},
		},
	}
}

func AwsStartEcsFailed(messageTs string, errorMessage string) slack.Msg {
	return slack.Msg{
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject(
						slack.MarkdownType,
						fmt.Sprintf(":x: *Failed to start ECS services*\n"+
							"Error: %s", errorMessage),
						false, false),
					nil, nil,
				),
			},
		},
	}
}