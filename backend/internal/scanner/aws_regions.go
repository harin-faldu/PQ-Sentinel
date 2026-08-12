package scanner

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// fallbackRegions is used only if DescribeRegions itself fails (e.g. the
// credentials can't even reach EC2). Kept intentionally broad, including
// ap-south-1 (Mumbai) which was missing from earlier drafts of this scanner.
var fallbackRegions = []string{
	"us-east-1", "us-east-2", "us-west-1", "us-west-2",
	"eu-west-1", "eu-west-2", "eu-central-1",
	"ap-south-1", "ap-south-2", "ap-southeast-1", "ap-southeast-2",
	"ap-northeast-1", "ap-northeast-2",
	"sa-east-1", "ca-central-1", "me-central-1",
}

// DiscoverEnabledRegions queries the account's actually-enabled regions.
// Falls back to a static list only if the API call itself fails — this is
// safer than hardcoding, since accounts vary in which regions are opted in.
func DiscoverEnabledRegions(ctx context.Context, cfg aws.Config) []string {
	client := ec2.NewFromConfig(cfg)
	out, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil || len(out.Regions) == 0 {
		return fallbackRegions
	}
	regions := make([]string, 0, len(out.Regions))
	for _, r := range out.Regions {
		if r.RegionName != nil {
			regions = append(regions, *r.RegionName)
		}
	}
	return regions
}
