package resources

import (
	"log/slog"
	"regexp"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/wasilak/cloudpile/cache"
)

type EC2Sg struct {
	EC2Svc *ec2.EC2
	BaseAWSResource
}

func (r *EC2Sg) Init(cache cache.Cache) error {
	return nil
}

func (r *EC2Sg) Get() ([]Item, error) {
	var items []Item
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

		item := Item{
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
