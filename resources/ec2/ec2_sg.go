package ec2

import (
	"log/slog"
	"regexp"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/wasilak/cloudpile/cache"
	"github.com/wasilak/cloudpile/resources"
)

type EC2Sg struct {
	EC2Svc *ec2.EC2
	resources.BaseAWSResource
}

func (r *EC2Sg) Init(cache cache.Cache) error {
	return nil
}

func (r *EC2Sg) Get() ([]resources.Item, error) {
	var items []resources.Item
	var err error
	var result *ec2.DescribeSecurityGroupsOutput

	result, err = r.EC2Svc.DescribeSecurityGroups(nil)
	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if !match {
			slog.Debug("Error", slog.AnyValue(err))
		}
		return items, err
	}

	for _, sg := range result.SecurityGroups {

		item := resources.Item{
			ID:           *sg.GroupId,
			Type:         "Security Group",
			Tags:         sg.Tags,
			Account:      r.AccountID,
			AccountAlias: r.AccountAlias,
			Region:       *r.EC2Svc.Config.Region,
		}

		items = append(items, item)
	}

	return items, nil
}
