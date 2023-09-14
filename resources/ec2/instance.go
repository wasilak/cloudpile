package ec2

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/wasilak/cloudpile/cache"
	"github.com/wasilak/cloudpile/resources"
)

type EC2Instance struct {
	Client *ec2.Client
	resources.BaseAWSResource
}

func (r *EC2Instance) Init(cache cache.Cache) error {
	return nil
}

func (r *EC2Instance) Get() ([]resources.Item, error) {
	items := []resources.Item{}
	var err error
	var result *ec2.DescribeInstancesOutput

	// Call to get detailed information on each instance
	result, err = r.Client.DescribeInstances(context.TODO(), nil)
	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if !match {
			slog.Debug("Error", slog.AnyValue(err))
		}
		return items, err
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {

			privateIP := ""

			if instance.PrivateIpAddress != nil {
				privateIP = *instance.PrivateIpAddress
			}

			tags := []resources.ItemTag{}
			for _, v := range instance.Tags {
				newTag := resources.ItemTag{
					Key:   *v.Key,
					Value: *v.Value,
				}

				tags = append(tags, newTag)
			}

			item := resources.Item{
				ID:             *instance.InstanceId,
				Type:           "EC2 instance",
				Tags:           tags,
				Account:        r.AccountID,
				AccountAlias:   r.AccountAlias,
				Region:         r.Region,
				IP:             privateIP,
				PrivateDNSName: *instance.PrivateDnsName,
			}

			items = append(items, item)
		}
	}

	return items, nil
}
