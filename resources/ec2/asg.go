package ec2

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/wasilak/cloudpile/resources"
)

type ASG struct {
	Client *autoscaling.Client
	resources.BaseAWSResource
}

func (r *ASG) GetCacheKey() string {
	return fmt.Sprintf("%s-%s-%s", r.AccountID, r.Region, r.Type)
}

func (r *ASG) Get() ([]resources.Item, error) {
	items := []resources.Item{}
	var err error
	var result *autoscaling.DescribeAutoScalingGroupsOutput

	result, err = r.Client.DescribeAutoScalingGroups(context.TODO(), nil)
	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if !match {
			slog.Debug("Error", "error", err)
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
