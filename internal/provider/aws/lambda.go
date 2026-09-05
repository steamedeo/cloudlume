package aws

import (
	"context"
	"strconv"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/steamedeo/cloudlume/internal/model"
)

func fetchLambda(ctx context.Context, cfg awssdk.Config, account model.Account, region string) []model.Resource {
	client := lambda.NewFromConfig(cfg)

	var resources []model.Resource
	paginator := lambda.NewListFunctionsPaginator(client, &lambda.ListFunctionsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return resources
		}
		for _, fn := range page.Functions {
			resources = append(resources, functionToResource(fn, account, region))
		}
	}
	return resources
}

func functionToResource(fn types.FunctionConfiguration, account model.Account, region string) model.Resource {
	name := awssdk.ToString(fn.FunctionName)

	// ListFunctions only reliably populates State around creation/config
	// updates — most steady-state functions report it empty. Fall back to
	// LastUpdateStatus (deployment outcome), then to a plain "active"
	// assumption, so a resource is never left with a blank status.
	status := string(fn.State)
	var health model.Health
	switch fn.State {
	case types.StateActive:
		health = model.HealthOK
	case types.StatePending:
		health = model.HealthWarn
	case types.StateFailed, types.StateInactive:
		health = model.HealthDown
	default:
		switch fn.LastUpdateStatus {
		case types.LastUpdateStatusSuccessful:
			status, health = "active", model.HealthOK
		case types.LastUpdateStatusInProgress:
			status, health = "updating", model.HealthWarn
		case types.LastUpdateStatusFailed:
			status, health = "update failed", model.HealthDown
		default:
			status, health = "active", model.HealthOK
		}
	}

	memory := ""
	if fn.MemorySize != nil {
		memory = strconv.Itoa(int(*fn.MemorySize)) + " MB"
	}
	timeout := ""
	if fn.Timeout != nil {
		timeout = strconv.Itoa(int(*fn.Timeout)) + "s"
	}

	details := []model.Detail{
		{Label: "Runtime", Value: string(fn.Runtime)},
		{Label: "Memory", Value: memory},
		{Label: "Timeout", Value: timeout},
		{Label: "Handler", Value: awssdk.ToString(fn.Handler)},
		{Label: "Last Modified", Value: awssdk.ToString(fn.LastModified)},
		{Label: "Last Deploy", Value: string(fn.LastUpdateStatus)},
	}

	return model.Resource{
		Provider: "aws",
		Account:  account.Name,
		Region:   region,
		Category: model.CategoryLambda,
		Type:     "lambda-function",
		ID:       name,
		Name:     name,
		Status:   status,
		Health:   health,
		Details:  details,
	}
}
