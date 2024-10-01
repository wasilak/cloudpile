package ec2

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/wasilak/cloudpile/resources"
)

type EC2Sg struct {
	Client *ec2.Client
	resources.BaseAWSResource
}

func (r *EC2Sg) GetCacheKey() string {
	return fmt.Sprintf("%s-%s-%s", r.AccountID, r.Region, r.Type)
}

func (r *EC2Sg) Get() ([]resources.Item, error) {
	var items []resources.Item
	var err error
	var result *ec2.DescribeSecurityGroupsOutput

	result, err = r.Client.DescribeSecurityGroups(context.TODO(), nil)
	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if !match {
			slog.Debug("Error", "error", err)
		}
		return items, err
	}

	for _, sg := range result.SecurityGroups {

		tags := []resources.ItemTag{}
		for _, v := range sg.Tags {
			newTag := resources.ItemTag{
				Key:   *v.Key,
				Value: *v.Value,
			}

			tags = append(tags, newTag)
		}

		item := resources.Item{
			ID:           *sg.GroupId,
			Type:         "Security Group",
			Tags:         tags,
			Account:      r.AccountID,
			AccountAlias: r.AccountAlias,
			Region:       r.Region,
		}

		items = append(items, item)
	}

	return items, nil
}
