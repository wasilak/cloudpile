package libs

import (
	"net"
	"slices"
	"sync"
	"time"

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

	chanItems := make(chan []resources.Item)

	var wg sync.WaitGroup

	for _, awsConfig := range AWSConfigs {
		for _, region := range awsConfig.Regions {

			awsConfigV2, err := newAWSV2Config(awsConfig, region)
			if err != nil {
				slog.Debug(err.Error(), "awsConfig", awsConfig, "region", region)
			} else {
				fetchItems(&wg, chanItems, region, awsConfigV2, awsConfig, cacheInstance, forceRefresh)
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

	if len(IDs) == 0 {
		return items, nil
	}

	filteredItems := filterItems(items, IDs)

	return filteredItems, nil
}

func fetchItems(wg *sync.WaitGroup, chanItems chan<- []resources.Item, region string, awsConfigV2 aws.Config, awsConfig AWSConfig, cacheInstance cache.Cache, forceRefresh bool) {
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
				Type:         "ec2",
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
				Type:         "sg",
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
				Type:         "lambda",
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
				Type:         "elb",
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
				Type:         "asg",
			},
		})
	}

	for _, itemsType := range res {
		wg.Add(1)
		go describeItems(wg, chanItems, cacheInstance, forceRefresh, itemsType)
	}
}

func describeItems(wg *sync.WaitGroup, chanItems chan<- []resources.Item, cacheInstance cache.Cache, forceRefresh bool, res resources.AWSResourceType) {
	defer wg.Done()
	var result interface{}
	var err error

	items := []resources.Item{}

	if cacheInstance.Enabled {

		var found bool

		result, found = cacheInstance.Cache.Get(res.GetCacheKey())

		if !forceRefresh && !found {
			slog.Debug("Cache not yet initialized", "cache_key", res.GetCacheKey(), "forceRefresh", forceRefresh)
		}

		if found {
			slog.Debug("Cache hit", "cache_key", res.GetCacheKey(), "forceRefresh", forceRefresh)
			items = result.([]resources.Item)
		} else {
			slog.Debug("Cache miss", "cache_key", res.GetCacheKey(), "forceRefresh", forceRefresh)
		}

		if forceRefresh {
			items, err = res.Get()
			if err != nil {
				slog.Error(err.Error())
			}

			// set a value with a cost of 1
			cacheInstance.Cache.Set(res.GetCacheKey(), items, 1)

			// wait for value to pass through buffers
			cacheInstance.Cache.Wait()

			nextUpdate := time.Now().Add(cacheInstance.TTL)

			slog.Debug("Cache refresh done", "cache_key", res.GetCacheKey(), "forceRefresh", forceRefresh, "next_in", cache.CacheInstance.TTL, "next_time", nextUpdate)
		}

	} else {
		items, err = res.Get()
		if err != nil {
			slog.Error(err.Error())
		}
	}

	chanItems <- items
}

func filterItems(items []resources.Item, IDs []string) []resources.Item {
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
