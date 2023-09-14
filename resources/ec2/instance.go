package ec2

import (
	"log/slog"
	"regexp"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/wasilak/cloudpile/cache"
	"github.com/wasilak/cloudpile/resources"
)

type EC2Instance struct {
	EC2Svc *ec2.EC2
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
	result, err = r.EC2Svc.DescribeInstances(nil)
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

			item := resources.Item{
				ID:             *instance.InstanceId,
				Type:           "EC2 instance",
				Tags:           instance.Tags,
				Account:        r.AccountID,
				AccountAlias:   r.AccountAlias,
				Region:         *r.EC2Svc.Config.Region,
				IP:             privateIP,
				PrivateDNSName: *instance.PrivateDnsName,
			}

			items = append(items, item)
		}
	}

	return items, nil
}
