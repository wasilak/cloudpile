package libs

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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
func Describe(awsRegions, IDs, iamRoles []string, accountAliasses map[string]string, verbose bool, cacheInstance Cache, forceRefresh bool) Items {

	items := Items{}
	filteredItems := Items{}
	var result interface{}

	if cacheInstance.Enabled {

		var found bool

		cacheKey := fmt.Sprintf("app_cache")

		result, found = cacheInstance.Cache.Get(cacheKey)

		if !forceRefresh && !found {
			log.Println("Cache not yet initialized")
			return items
		}

		if found {
			items = result.(Items)
		}

		if forceRefresh {
			items = refreshCache(awsRegions, iamRoles, accountAliasses, verbose, cacheInstance, forceRefresh, cacheKey)
		}

	}

	if len(IDs) == 0 {
		return items
	}

	filteredItems = append(filteredItems, filterEc2(items, IDs)...)

	filteredItems = append(filteredItems, filterSg(items, IDs)...)

	return filteredItems
}

func refreshCache(awsRegions, iamRoles []string, accountAliasses map[string]string, verbose bool, cacheInstance Cache, forceRefresh bool, cacheKey string) Items {
	var sess *session.Session

	items := Items{}

	var wg sync.WaitGroup

	itemsChannel := make(chan Items)

	// this is required, in order to prevent following situation:
	// 1. goroutine runs in loop, pushing result to channel
	// 2. application pauses waiting for result to be processed (taken from channel)
	// thanks to this, `range itemsChannel` is able to process items as they appear in channel
	// thus, unblocking goroutine processing :)
	// see: https://dev.to/sophiedebenedetto/synchronizing-go-routines-with-channels-and-waitgroups-3ke2
	go func() {
		wg.Wait()
		close(itemsChannel)
	}()

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

			go runDescribe(&wg, creds, itemsChannel, sess, region, accountID, accountAlias, verbose, cacheInstance, forceRefresh)
		}
	}

	for item := range itemsChannel {
		items = append(items, item...)
	}

	// set a value with a cost of 1
	cacheInstance.Cache.Set(cacheKey, items, 1)

	// wait for value to pass through buffers
	cacheInstance.Cache.Wait()

	return items
}

func runDescribe(wg *sync.WaitGroup, creds *credentials.Credentials, itemsChannel chan Items, sess *session.Session, region, accountID, accountAlias string, verbose bool, cacheInstance Cache, forceRefresh bool) {
	defer wg.Done()

	awsRegion := aws.String(region)

	// Create new EC2 client
	ec2Svc := ec2.New(sess, &aws.Config{
		Credentials: creds,
		Region:      awsRegion,
	})

	itemsChannel <- describeEc2(ec2Svc, verbose, accountID, accountAlias, awsRegion)

	itemsChannel <- describeSg(ec2Svc, verbose, accountID, accountAlias, awsRegion)
}

func describeSg(ec2Svc *ec2.EC2, verbose bool, account, accountAlias string, awsRegion *string) Items {
	var items []Item
	var err error
	var result *ec2.DescribeSecurityGroupsOutput

	result, err = ec2Svc.DescribeSecurityGroups(nil)
	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if verbose || !match {
			log.Println("Error", err)
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

func describeEc2(ec2Svc *ec2.EC2, verbose bool, account, accountAlias string, awsRegion *string) Items {
	var items []Item
	var err error
	var result *ec2.DescribeInstancesOutput

	// Call to get detailed information on each instance
	result, err = ec2Svc.DescribeInstances(nil)
	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if verbose || !match {
			log.Println("Error", err)
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
