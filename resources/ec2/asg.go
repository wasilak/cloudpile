package ec2

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/wasilak/cloudpile/cache"
	"github.com/wasilak/cloudpile/resources"
)

type ASG struct {
	Client *autoscaling.Client
	resources.BaseAWSResource
}

func (r *ASG) Init(cache cache.Cache) error {
	return nil
}

func (r *ASG) Get() ([]resources.Item, error) {
	items := []resources.Item{}
	var err error
	var result *autoscaling.DescribeAutoScalingGroupsOutput

	result, err = r.Client.DescribeAutoScalingGroups(context.TODO(), nil)
	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if !match {
			slog.Debug("Error", slog.AnyValue(err))
		}
		return items, err
	}

	for _, item := range result.AutoScalingGroups {

		tags := []resources.ItemTag{}
		for _, tag := range item.Tags {
			newTag := resources.ItemTag{
				Key:   *tag.Key,
				Value: *tag.Value,
			}

			tags = append(tags, newTag)
		}

		item := resources.Item{
			Type:         "AutoScaling group",
			ARN:          *item.AutoScalingGroupARN,
			Tags:         tags,
			Account:      r.AccountID,
			AccountAlias: r.AccountAlias,
			Region:       r.Region,
		}

		items = append(items, item)
	}

	return items, nil
}
