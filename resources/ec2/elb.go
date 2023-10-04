package ec2

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/wasilak/cloudpile/resources"
)

type ELB struct {
	Client *elasticloadbalancingv2.Client
	resources.BaseAWSResource
}

func (r *ELB) GetCacheKey() string {
	return fmt.Sprintf("%s-%s-%s", r.AccountID, r.Region, r.Type)
}

func (r *ELB) Get() ([]resources.Item, error) {
	items := []resources.Item{}
	var err error
	var result *elasticloadbalancingv2.DescribeLoadBalancersOutput

	result, err = r.Client.DescribeLoadBalancers(context.TODO(), nil)
	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if !match {
			slog.Debug("Error", slog.AnyValue(err))
		}
		return items, err
	}

	for _, item := range result.LoadBalancers {

		describeTagsInput := elasticloadbalancingv2.DescribeTagsInput{
			ResourceArns: []string{
				*item.LoadBalancerArn,
			},
		}

		tagsOutput, err := r.Client.DescribeTags(context.TODO(), &describeTagsInput, func(*elasticloadbalancingv2.Options) {})
		if err != nil {
			return items, err
		}

		tags := []resources.ItemTag{}
		for _, item := range tagsOutput.TagDescriptions {
			for _, tag := range item.Tags {
				newTag := resources.ItemTag{
					Key:   *tag.Key,
					Value: *tag.Value,
				}

				tags = append(tags, newTag)
			}
		}

		item := resources.Item{
			Type:           fmt.Sprintf("ELB (%s)", item.Type),
			Tags:           tags,
			Account:        r.AccountID,
			AccountAlias:   r.AccountAlias,
			Region:         r.Region,
			PrivateDNSName: *item.DNSName,
		}

		items = append(items, item)
	}

	return items, nil
}
