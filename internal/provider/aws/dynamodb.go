package aws

import (
	"context"
	"strconv"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/steamedeo/cloudlume/internal/model"
)

func fetchDynamoDB(ctx context.Context, cfg awssdk.Config, account model.Account, region string) []model.Resource {
	client := dynamodb.NewFromConfig(cfg)

	var resources []model.Resource
	paginator := dynamodb.NewListTablesPaginator(client, &dynamodb.ListTablesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return resources
		}
		for _, name := range page.TableNames {
			out, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: awssdk.String(name)})
			if err != nil || out.Table == nil {
				continue
			}
			resources = append(resources, tableToResource(*out.Table, account, region))
		}
	}
	return resources
}

func tableToResource(t types.TableDescription, account model.Account, region string) model.Resource {
	name := awssdk.ToString(t.TableName)
	status := string(t.TableStatus)

	health := model.HealthUnknown
	switch t.TableStatus {
	case types.TableStatusActive:
		health = model.HealthOK
	case types.TableStatusCreating, types.TableStatusUpdating:
		health = model.HealthWarn
	case types.TableStatusDeleting, types.TableStatusInaccessibleEncryptionCredentials, types.TableStatusArchiving:
		health = model.HealthDown
	}

	billingMode := "PROVISIONED"
	if t.BillingModeSummary != nil {
		billingMode = string(t.BillingModeSummary.BillingMode)
	}

	capacity := ""
	if billingMode == string(types.BillingModePayPerRequest) {
		capacity = "on-demand"
	} else if t.ProvisionedThroughput != nil {
		read := awssdk.ToInt64(t.ProvisionedThroughput.ReadCapacityUnits)
		write := awssdk.ToInt64(t.ProvisionedThroughput.WriteCapacityUnits)
		capacity = strconv.FormatInt(read, 10) + " RCU / " + strconv.FormatInt(write, 10) + " WCU"
	}

	itemCount := ""
	if t.ItemCount != nil {
		itemCount = strconv.FormatInt(*t.ItemCount, 10)
	}

	size := ""
	if t.TableSizeBytes != nil {
		size = strconv.FormatInt(*t.TableSizeBytes/(1024*1024), 10) + " MB"
	}

	details := []model.Detail{
		{Label: "Billing Mode", Value: billingMode},
		{Label: "Capacity", Value: capacity},
		{Label: "Item Count", Value: itemCount},
		{Label: "Size", Value: size},
		{Label: "Created", Value: formatTime(t.CreationDateTime)},
	}

	return model.Resource{
		Provider: "aws",
		Account:  account.Name,
		Region:   region,
		Category: model.CategoryDatabase,
		Type:     "dynamodb-table",
		ID:       name,
		Name:     name,
		Status:   status,
		Health:   health,
		Details:  details,
	}
}
