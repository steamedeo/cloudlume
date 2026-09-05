package aws

import (
	"context"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"

	"github.com/steamedeo/cloudlume/internal/model"
)

// fetchNetworking lists CloudFront distributions, which — like S3 — is a
// global (partition-wide) API rather than per-region.
func fetchNetworking(ctx context.Context, cfg awssdk.Config, account model.Account) []model.Resource {
	client := cloudfront.NewFromConfig(cfg)

	var resources []model.Resource
	paginator := cloudfront.NewListDistributionsPaginator(client, &cloudfront.ListDistributionsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return resources
		}
		if page.DistributionList == nil {
			continue
		}
		for _, d := range page.DistributionList.Items {
			resources = append(resources, distributionToResource(d, account))
		}
	}
	return resources
}

func distributionToResource(d types.DistributionSummary, account model.Account) model.Resource {
	id := awssdk.ToString(d.Id)

	name := awssdk.ToString(d.DomainName)
	if d.Aliases != nil && len(d.Aliases.Items) > 0 {
		name = strings.Join(d.Aliases.Items, ", ")
	}
	if name == "" {
		name = id
	}

	status := awssdk.ToString(d.Status)
	health := model.HealthUnknown
	switch {
	case d.Enabled != nil && !*d.Enabled:
		health = model.HealthWarn
	case status == "Deployed":
		health = model.HealthOK
	case status == "InProgress":
		health = model.HealthWarn
	}

	enabled := "no"
	if d.Enabled != nil && *d.Enabled {
		enabled = "yes"
	}

	details := []model.Detail{
		{Label: "Domain", Value: awssdk.ToString(d.DomainName)},
		{Label: "Enabled", Value: enabled},
		{Label: "Price Class", Value: string(d.PriceClass)},
		{Label: "HTTP Version", Value: string(d.HttpVersion)},
		{Label: "Comment", Value: awssdk.ToString(d.Comment)},
		{Label: "Last Modified", Value: formatTime(d.LastModifiedTime)},
	}

	return model.Resource{
		Provider: "aws",
		Account:  account.Name,
		Region:   "global",
		Category: model.CategoryNetwork,
		Type:     "cloudfront-distribution",
		ID:       id,
		Name:     name,
		Status:   status,
		Health:   health,
		Details:  details,
	}
}
