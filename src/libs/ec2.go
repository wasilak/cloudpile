package libs

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"net"
	"regexp"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// Item type
type Item struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"`
	Tags         []*ec2.Tag `json:"tags"`
	Account      string     `json:"account"`
	AccountAlias string     `json:"accountAlias"`
	Region       string     `json:"region"`
	IP           string     `json:"ip"`
}

// Items type
type Items []Item

// Describe func
func Describe(awsRegions, IDs, iamRoles []string, accountAliasses map[string]string, cacheInstance Cache, forceRefresh bool, txn *newrelic.Transaction) Items {
	defer txn.StartSegment("libs.Describe").End()

	items := Items{}
	filteredItems := Items{}
	var result interface{}

	if cacheInstance.Enabled {

		var found bool

		cacheKey := fmt.Sprintf("app_cache")

		result, found = cacheInstance.Cache.Get(cacheKey)

		if !forceRefresh && !found {
			log.Debug("Cache not yet initialized")
			return items
		}

		if found {
			items = result.(Items)
		}

		if forceRefresh {
			items = refreshCache(awsRegions, iamRoles, accountAliasses, cacheInstance, forceRefresh, cacheKey, txn)
		}

	}

	if len(IDs) == 0 {
		return items
	}

	filteredItems = append(filteredItems, filterEc2(items, IDs)...)

	filteredItems = append(filteredItems, filterSg(items, IDs)...)

	return filteredItems
}

func refreshCache(awsRegions, iamRoles []string, accountAliasses map[string]string, cacheInstance Cache, forceRefresh bool, cacheKey string, txn *newrelic.Transaction) Items {
	defer txn.StartSegment("libs.refreshCache").End()

	var sess *session.Session

	items := Items{}

	var wg sync.WaitGroup

	for _, iamRole := range iamRoles {
		for _, region := range awsRegions {
			wg.Add(1)
			sess = session.Must(session.NewSession())
			creds := stscreds.NewCredentials(sess, iamRole)
			accountID := getAccountIdFromRoleARN(iamRole)

			accountAlias := ""

			if val, ok := accountAliasses[accountID]; ok {
				accountAlias = val
			}

			newItems := runDescribe(&wg, creds, sess, region, accountID, accountAlias, cacheInstance, forceRefresh, txn)

			items = append(items, newItems...)
		}
	}

	wg.Wait()

	// set a value with a cost of 1
	cacheInstance.Cache.Set(cacheKey, items, 1)

	// wait for value to pass through buffers
	cacheInstance.Cache.Wait()

	return items
}

func runDescribe(wg *sync.WaitGroup, creds *credentials.Credentials, sess *session.Session, region, accountID, accountAlias string, cacheInstance Cache, forceRefresh bool, txn *newrelic.Transaction) Items {
	defer txn.StartSegment("libs.runDescribe").End()

	defer wg.Done()

	items := Items{}

	awsRegion := aws.String(region)

	// Create new EC2 client
	ec2Svc := ec2.New(sess, &aws.Config{
		Credentials: creds,
		Region:      awsRegion,
	})

	items = append(items, describeEc2(ec2Svc, accountID, accountAlias, awsRegion, txn)...)
	items = append(items, describeSg(ec2Svc, accountID, accountAlias, awsRegion, txn)...)

	return items
}

func describeSg(ec2Svc *ec2.EC2, account, accountAlias string, awsRegion *string, txn *newrelic.Transaction) Items {
	defer txn.StartSegment("libs.describeSg").End()

	var items []Item
	var err error
	var result *ec2.DescribeSecurityGroupsOutput

	result, err = ec2Svc.DescribeSecurityGroups(nil)
	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if !match {
			log.Debug("Error", err)
		}
	}

	for _, sg := range result.SecurityGroups {

		var tags []*ec2.Tag
		for _, tag := range sg.Tags {
			tags = append(tags, tag)
		}

		item := Item{
			ID:           *sg.GroupId,
			Type:         "Security Group",
			Tags:         tags,
			Account:      account,
			AccountAlias: accountAlias,
			Region:       *awsRegion,
		}

		items = append(items, item)
	}

	return items
}

func describeEc2(ec2Svc *ec2.EC2, account, accountAlias string, awsRegion *string, txn *newrelic.Transaction) Items {
	defer txn.StartSegment("libs.describeEc2").End()

	var items []Item
	var err error
	var result *ec2.DescribeInstancesOutput

	// Call to get detailed information on each instance
	result, err = ec2Svc.DescribeInstances(nil)
	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if !match {
			log.Debug("Error", err)
		}
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {

			privateIP := ""

			if instance.PrivateIpAddress != nil {
				privateIP = *instance.PrivateIpAddress
			}

			var tags []*ec2.Tag
			for _, tag := range instance.Tags {
				tags = append(tags, tag)
			}

			item := Item{
				ID:           *instance.InstanceId,
				Type:         "EC2 instance",
				Tags:         tags,
				Account:      account,
				AccountAlias: accountAlias,
				Region:       *ec2Svc.Config.Region,
				IP:           privateIP,
			}

			items = append(items, item)
		}
	}

	return items
}

func filterEc2(items Items, IDs []string) Items {
	var filteredItems []Item
	var resourceIDs []string
	var resourceIPs []string
	var match bool

	for _, id := range IDs {
		// EC2 instances
		match, _ = regexp.MatchString("i-[a-zA-Z0-9_]+", id)
		if match {
			resourceIDs = append(resourceIDs, id)
		}

		if net.ParseIP(id) != nil {
			resourceIPs = append(resourceIPs, id)
		}
	}

	if len(IDs) != 0 && len(resourceIDs) == 0 && len(resourceIPs) == 0 {
		return Items{}
	}

	for _, item := range items {

		hit := false

		for _, id := range resourceIDs {
			if item.ID == id {
				hit = true
			}

			if hit == true {
				break
			}
		}

		for _, ip := range resourceIPs {
			if item.IP == ip {
				hit = true
			}

			if hit == true {
				break
			}
		}

		if hit == false {
			continue
		}

		filteredItems = append(filteredItems, item)
	}

	return filteredItems
}

func filterSg(items Items, IDs []string) Items {
	var filteredItems []Item
	var resourceIDs []string
	var match bool

	for _, id := range IDs {
		match, _ = regexp.MatchString("sg-[a-zA-Z0-9_]+", id)
		if match {
			resourceIDs = append(resourceIDs, id)
		}
	}

	if len(IDs) != 0 && len(resourceIDs) == 0 {
		return Items{}
	}

	for _, item := range items {
		for _, id := range resourceIDs {
			if item.ID == id {
				filteredItems = append(filteredItems, item)
				break
			}
		}
	}

	return filteredItems
}
