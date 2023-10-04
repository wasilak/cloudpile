package libs

import (
	"net"
	"slices"
	"sync"

	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/wasilak/cloudpile/cache"
	"github.com/wasilak/cloudpile/resources"
	ec2Resource "github.com/wasilak/cloudpile/resources/ec2"
)

func Run(IDs []string, cacheInstance cache.Cache, forceRefresh bool) ([]resources.Item, error) {
	items := []resources.Item{}
	filteredItems := []resources.Item{}
	var result interface{}
	var err error

	if cacheInstance.Enabled {

		var found bool

		cacheKey := "app_cache"

		result, found = cacheInstance.Cache.Get(cacheKey)

		if !forceRefresh && !found {
			slog.Debug("Cache not yet initialized")
			return items, nil
		}

		if found {
			slog.Debug("Cache hit...")
			items = result.([]resources.Item)
		} else {
			slog.Debug("Cache miss...")
		}

		if forceRefresh {
			items, err = refreshCache(cacheInstance, forceRefresh, cacheKey)
			if err != nil {
				return nil, err
			}
		}

	} else {
		items, err = getItems()
		if err != nil {
			return nil, err
		}
	}

	if len(IDs) == 0 {
		return items, nil
	}

	filteredItems = append(filteredItems, filterEc2(items, IDs)...)

	return filteredItems, nil
}

func refreshCache(cacheInstance cache.Cache, forceRefresh bool, cacheKey string) ([]resources.Item, error) {
	items, err := getItems()
	if err != nil {
		return nil, nil
	}

	// set a value with a cost of 1
	cacheInstance.Cache.Set(cacheKey, items, 1)

	// wait for value to pass through buffers
	cacheInstance.Cache.Wait()

	return items, nil
}

func getItems() ([]resources.Item, error) {
	chanItems := make(chan []resources.Item)

	var wg sync.WaitGroup

	for _, awsConfig := range AWSConfigs {
		for _, region := range awsConfig.Regions {

			awsConfigV2, err := newAWSV2Config(awsConfig, region)
			if err != nil {
				slog.Debug(err.Error())
			} else {
				fetchItems(&wg, chanItems, region, awsConfigV2, awsConfig)
			}
		}
	}

	go func() {
		wg.Wait()
		close(chanItems)
	}()

	items := []resources.Item{}
	for result := range chanItems {
		items = append(items, result...)
	}

	return items, nil
}

func fetchItems(wg *sync.WaitGroup, chanItems chan<- []resources.Item, region string, awsConfigV2 aws.Config, awsConfig AWSConfig) {
	res := []resources.AWSResourceType{}
	var (
		accountID string
		err       error
	)

	accountID, err = getAccountId(awsConfigV2)
	if err != nil {
		slog.Error(err.Error())
		accountID = ""
	}

	ec2Client := ec2.NewFromConfig(awsConfigV2)

	// EC2 instances
	if slices.Contains(awsConfig.Resources, "ec2") {

		res = append(res, &ec2Resource.EC2Instance{
			Client: ec2Client,
			BaseAWSResource: resources.BaseAWSResource{
				AccountID:    accountID,
				AccountAlias: awsConfig.AccountAlias,
				Region:       region,
			},
		})
	}

	// EC2 security groups
	if slices.Contains(awsConfig.Resources, "sg") {
		res = append(res, &ec2Resource.EC2Sg{
			Client: ec2Client,
			BaseAWSResource: resources.BaseAWSResource{
				AccountID:    accountID,
				AccountAlias: awsConfig.AccountAlias,
				Region:       region,
			},
		})
	}

	// Lambda functions
	if slices.Contains(awsConfig.Resources, "lambda") {

		lambdaClient := lambda.NewFromConfig(awsConfigV2)

		res = append(res, &resources.LambdaFunction{
			Client: lambdaClient,
			BaseAWSResource: resources.BaseAWSResource{
				AccountID:    accountID,
				AccountAlias: awsConfig.AccountAlias,
				Region:       region,
			},
		})
	}

	// EC2 load balancers
	if slices.Contains(awsConfig.Resources, "elb") {

		elbClient := elasticloadbalancingv2.NewFromConfig(awsConfigV2)

		res = append(res, &ec2Resource.ELB{
			Client: elbClient,
			BaseAWSResource: resources.BaseAWSResource{
				AccountID:    accountID,
				AccountAlias: awsConfig.AccountAlias,
				Region:       region,
			},
		})
	}

	// EC2 autoscaling groups
	if slices.Contains(awsConfig.Resources, "asg") {

		asgClient := autoscaling.NewFromConfig(awsConfigV2)

		res = append(res, &ec2Resource.ASG{
			Client: asgClient,
			BaseAWSResource: resources.BaseAWSResource{
				AccountID:    accountID,
				AccountAlias: awsConfig.AccountAlias,
				Region:       region,
			},
		})
	}

	for _, v := range res {
		wg.Add(1)
		go describeItems(wg, chanItems, v)
	}
}

func describeItems(wg *sync.WaitGroup, chanItems chan<- []resources.Item, res resources.AWSResourceType) {
	defer wg.Done()

	items, err := res.Get()
	if err != nil {
		slog.Error(err.Error())
	}

	chanItems <- items
}

func filterEc2(items []resources.Item, IDs []string) []resources.Item {
	var filteredItems []resources.Item
	var resourceIDs []string
	var resourceIPs []string
	var resourceTags []map[string]string

	for _, id := range IDs {
		// tags
		tags := getTagsFromString(id)
		if len(tags) > 0 {
			resourceTags = append(resourceTags, tags)
		}

		// IP is kinda special as it is not string, everything else can be matched in loop below
		if net.ParseIP(id) != nil {
			resourceIPs = append(resourceIPs, id)
		} else {
			resourceIDs = append(resourceIDs, id)
		}

	}

	if len(IDs) != 0 && len(resourceIDs) == 0 && len(resourceIPs) == 0 {
		return []resources.Item{}
	}

	for _, item := range items {

		hit := false

		for _, id := range resourceIDs {
			if item.ID == id || item.PrivateDNSName == id {
				hit = true
			}

			if hit {
				break
			}
		}

		for _, id := range resourceIDs {
			if item.ARN == id || item.PrivateDNSName == id {
				hit = true
			}

			if hit {
				break
			}
		}

		for _, ip := range resourceIPs {
			if item.IP == ip {
				hit = true
			}

			if hit {
				break
			}
		}

		tagHit := 0
		for _, v := range resourceTags {
			for _, itemTag := range item.Tags {
				if itemTag.Key == v["name"] && itemTag.Value == v["value"] {
					tagHit++
				}
			}
		}

		if len(resourceTags) > 0 && tagHit == len(resourceTags) {
			hit = true
		}

		if !hit {
			continue
		}

		filteredItems = append(filteredItems, item)
	}

	return filteredItems
}
