package aws

import (
	"context"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/steamedeo/cloudlume/internal/model"
)

func fetchCompute(ctx context.Context, cfg awssdk.Config, account model.Account, region string) []model.Resource {
	client := ec2.NewFromConfig(cfg)

	var resources []model.Resource
	paginator := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return resources
		}
		for _, reservation := range page.Reservations {
			for _, inst := range reservation.Instances {
				resources = append(resources, instanceToResource(inst, account, region))
			}
		}
	}
	return resources
}

func instanceToResource(inst types.Instance, account model.Account, region string) model.Resource {
	name := ""
	for _, tag := range inst.Tags {
		if awssdk.ToString(tag.Key) == "Name" {
			name = awssdk.ToString(tag.Value)
		}
	}
	id := awssdk.ToString(inst.InstanceId)
	if name == "" {
		name = id
	}

	status := ""
	health := model.HealthUnknown
	if inst.State != nil {
		status = string(inst.State.Name)
		switch inst.State.Name {
		case types.InstanceStateNameRunning:
			health = model.HealthOK
		case types.InstanceStateNamePending, types.InstanceStateNameStopping, types.InstanceStateNameShuttingDown:
			health = model.HealthWarn
		case types.InstanceStateNameStopped, types.InstanceStateNameTerminated:
			health = model.HealthDown
		}
	}

	details := []model.Detail{
		{Label: "Instance Type", Value: string(inst.InstanceType)},
		{Label: "Availability Zone", Value: awssdk.ToString(placementAZ(inst))},
		{Label: "Public IP", Value: awssdk.ToString(inst.PublicIpAddress)},
		{Label: "Private IP", Value: awssdk.ToString(inst.PrivateIpAddress)},
		{Label: "VPC", Value: awssdk.ToString(inst.VpcId)},
		{Label: "Launched", Value: formatTime(inst.LaunchTime)},
	}

	return model.Resource{
		Provider: "aws",
		Account:  account.Name,
		Region:   region,
		Category: model.CategoryCompute,
		Type:     "ec2-instance",
		ID:       id,
		Name:     name,
		Status:   status,
		Health:   health,
		Details:  details,
	}
}

func placementAZ(inst types.Instance) *string {
	if inst.Placement != nil {
		return inst.Placement.AvailabilityZone
	}
	empty := ""
	return &empty
}
