package aws

import (
	"context"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/steamedeo/cloudlume/internal/model"
)

// fetchStorage lists S3 buckets, which is a global (partition-wide) API
// call rather than per-region. Each bucket's own region is looked up
// separately so the resource carries an accurate region label.
func fetchStorage(ctx context.Context, cfg awssdk.Config, account model.Account) []model.Resource {
	client := s3.NewFromConfig(cfg)

	out, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil
	}

	resources := make([]model.Resource, 0, len(out.Buckets))
	for _, b := range out.Buckets {
		resources = append(resources, bucketToResource(ctx, client, b, account))
	}
	return resources
}

func bucketToResource(ctx context.Context, client *s3.Client, b types.Bucket, account model.Account) model.Resource {
	name := awssdk.ToString(b.Name)

	region := "us-east-1"
	if loc, err := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{Bucket: b.Name}); err == nil {
		if r := string(loc.LocationConstraint); r != "" {
			region = r
		}
	}

	details := []model.Detail{
		{Label: "Created", Value: formatTime(b.CreationDate)},
	}

	return model.Resource{
		Provider: "aws",
		Account:  account.Name,
		Region:   region,
		Category: model.CategoryStorage,
		Type:     "s3-bucket",
		ID:       name,
		Name:     name,
		Status:   "available",
		Health:   model.HealthOK,
		Details:  details,
	}
}
