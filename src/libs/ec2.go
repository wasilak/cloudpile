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
func Describe(awsRegions, IDs, iamRoles []string, sess *session.Session, accountAliasses map[string]string, verbose bool, cacheInstance Cache) Items {

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
			creds := stscreds.NewCredentials(sess, iamRole)
			accountID := getAccountIdFromRoleARN(iamRole)

			accountAlias := ""

			if val, ok := accountAliasses[accountID]; ok {
				accountAlias = val
			}

			go runDescribe(&wg, creds, itemsChannel, sess, region, accountID, accountAlias, IDs, verbose, cacheInstance)
		}
	}

	for item := range itemsChannel {
		items = append(items, item...)
	}

	// if verbose {
	// 	log.Println(items)
	// }

	return items
}

func runDescribe(wg *sync.WaitGroup, creds *credentials.Credentials, itemsChannel chan Items, sess *session.Session, region, accountID, accountAlias string, IDs []string, verbose bool, cacheInstance Cache) {
	defer wg.Done()

	awsRegion := aws.String(region)

	// Create new EC2 client
	ec2Svc := ec2.New(sess, &aws.Config{
		Credentials: creds,
		Region:      awsRegion,
	})

	items := describeEc2(ec2Svc, IDs, accountID, accountAlias, verbose, awsRegion, cacheInstance)
	itemsChannel <- items

	// items = describeSg(ec2Svc, IDs, accountID, accountAlias, verbose, cacheInstance)
	// itemsChannel <- items

	// for _, id := range IDs {

	// 	var match bool

	// 	// EC2 instances
	// 	match, _ = regexp.MatchString("i-[a-zA-Z0-9_]+", id)
	// 	if match {
	// 		items := describeEc2(ec2Svc, id, accountID, accountAlias, verbose)
	// 		itemsChannel <- items
	// 	}

	// 	// Security Groups
	// 	match, _ = regexp.MatchString("sg-[a-zA-Z0-9_]+", id)
	// 	if match {
	// 		items := describeSg(ec2Svc, awsID, accountID, accountAlias, verbose)
	// 		itemsChannel <- items
	// 	}

	// }
}

// func describeSg(ec2Svc *ec2.EC2, IDs []string, account, accountAlias string, verbose bool, cacheInstance Cache) Items {
// 	var items []Item

// 	input := &ec2.DescribeSecurityGroupsInput{
// 		GroupIds: IDs,
// 	}

// 	// Call to get detailed information on each instance
// 	result, err := ec2Svc.DescribeSecurityGroups(input)
// 	if err != nil {
// 		match, _ := regexp.MatchString("does not exist", err.Error())
// 		if verbose || !match {
// 			log.Println("Error", err)
// 		}
// 		return items
// 	}

// 	for _, sg := range result.SecurityGroups {

// 		var tags []*ec2.Tag
// 		for _, tag := range sg.Tags {
// 			tags = append(tags, tag)
// 		}

// 		item := Item{
// 			ID:           *sg.GroupId,
// 			Type:         "Security Group",
// 			Tags:         tags,
// 			Account:      account,
// 			AccountAlias: accountAlias,
// 			Region:       *ec2Svc.Config.Region,
// 		}

// 		items = append(items, item)
// 	}

// 	return items
// }

func recoverFullName(instance *ec2.Instance) {
	if r := recover(); r != nil {
		fmt.Println("recovered from ", r)
		log.Printf("%+v", instance)
	}
}

func describeEc2(ec2Svc *ec2.EC2, id []string, account, accountAlias string, verbose bool, awsRegion *string, cacheInstance Cache) Items {
	var items []Item
	var resourceIDs []string
	var resourceIPs []string
	var match bool
	var found bool
	var err error

	var result *ec2.DescribeInstancesOutput

	cacheKey := fmt.Sprintf("list_ec2_%s_%s", accountAlias, *awsRegion)

	for _, id := range id {
		// EC2 instances
		match, _ = regexp.MatchString("i-[a-zA-Z0-9_]+", id)
		if match {
			resourceIDs = append(resourceIDs, id)
		}

		if net.ParseIP(id) != nil {
			resourceIPs = append(resourceIPs, id)
		}
	}

	if cacheInstance.Enabled {

		var resultTmp interface{}

		resultTmp, found = cacheInstance.Cache.Get(cacheKey)

		if !found {
			result, err = ec2Svc.DescribeInstances(nil)

			if err != nil {
				match, _ := regexp.MatchString("does not exist", err.Error())
				if verbose || !match {
					log.Println("Error", err)
				}
			}

			// set a value with a cost of 1
			cacheInstance.Cache.SetWithTTL(cacheKey, result, 1, cacheInstance.TTL)

			// wait for value to pass through buffers
			cacheInstance.Cache.Wait()
		} else {
			result = resultTmp.(*ec2.DescribeInstancesOutput)
		}

	} else {
		// Call to get detailed information on each instance
		result, err = ec2Svc.DescribeInstances(nil)
		if err != nil {
			match, _ := regexp.MatchString("does not exist", err.Error())
			if verbose || !match {
				log.Println("Error", err)
			}
		}
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {

			privateIP := ""

			if instance.PrivateIpAddress != nil {
				privateIP = *instance.PrivateIpAddress
			}

			// defer recoverFullName(instance)
			if len(resourceIDs) > 0 || len(resourceIPs) > 0 {
				hit := false

				for _, id := range resourceIDs {
					if *instance.InstanceId == id {
						hit = true
					}

					if hit == true {
						break
					}
				}

				for _, ip := range resourceIPs {
					if privateIP == ip {
						hit = true
					}

					if hit == true {
						break
					}
				}

				if hit == false {
					continue
				}
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

// func listEc2(ec2Svc *ec2.EC2, account, accountAlias string, verbose bool) Items {
// 	var instanceIDs []*string
// 	return describeEc2(ec2Svc, instanceIDs, account, accountAlias, verbose)
// }
