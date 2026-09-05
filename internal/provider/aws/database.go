package aws

import (
	"context"
	"strconv"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"

	"github.com/steamedeo/cloudlume/internal/model"
)

func fetchDatabases(ctx context.Context, cfg awssdk.Config, account model.Account, region string) []model.Resource {
	client := rds.NewFromConfig(cfg)

	var resources []model.Resource
	paginator := rds.NewDescribeDBInstancesPaginator(client, &rds.DescribeDBInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return resources
		}
		for _, db := range page.DBInstances {
			resources = append(resources, dbInstanceToResource(db, account, region))
		}
	}
	return resources
}

func dbInstanceToResource(db types.DBInstance, account model.Account, region string) model.Resource {
	id := awssdk.ToString(db.DBInstanceIdentifier)
	status := awssdk.ToString(db.DBInstanceStatus)

	health := model.HealthUnknown
	switch status {
	case "available":
		health = model.HealthOK
	case "creating", "modifying", "backing-up", "starting", "stopping", "upgrading":
		health = model.HealthWarn
	case "stopped", "failed", "deleting", "incompatible-restore":
		health = model.HealthDown
	}

	multiAZ := "no"
	if db.MultiAZ != nil && *db.MultiAZ {
		multiAZ = "yes"
	}

	storage := ""
	if db.AllocatedStorage != nil {
		storage = strconv.Itoa(int(*db.AllocatedStorage)) + " GiB"
	}

	details := []model.Detail{
		{Label: "Engine", Value: awssdk.ToString(db.Engine) + " " + awssdk.ToString(db.EngineVersion)},
		{Label: "Instance Class", Value: awssdk.ToString(db.DBInstanceClass)},
		{Label: "Multi-AZ", Value: multiAZ},
		{Label: "Storage", Value: storage},
		{Label: "Endpoint", Value: endpointAddress(db)},
	}

	return model.Resource{
		Provider: "aws",
		Account:  account.Name,
		Region:   region,
		Category: model.CategoryDatabase,
		Type:     "rds-instance",
		ID:       id,
		Name:     id,
		Status:   status,
		Health:   health,
		Details:  details,
	}
}

func endpointAddress(db types.DBInstance) string {
	if db.Endpoint != nil {
		return awssdk.ToString(db.Endpoint.Address)
	}
	return ""
}
