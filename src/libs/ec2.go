package libs

import (
	"log"
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
	ID   string
	Type string
	Tags []*ec2.Tag
	Account string
	AccountAlias string
}

// Items type
type Items []Item

// Describe func
func Describe(awsRegions,  IDs, iamRoles []string, sess *session.Session, accountAliasses map[string]string, verbose bool) Items {

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
			
			go runDescribe(&wg, creds, itemsChannel, sess, region, accountID, accountAlias, IDs, verbose)
		}
	}

	for item := range itemsChannel {
		items = append(items, item...)
	}

	return items
}

func runDescribe(wg *sync.WaitGroup, creds *credentials.Credentials,itemsChannel chan Items, sess *session.Session, region, accountID, accountAlias string, IDs []string, verbose bool) {
	defer wg.Done()

	awsRegion := aws.String(region)

	for _, id := range IDs {
		awsID := aws.String(id)

		var match bool

		// Create new EC2 client
		ec2Svc := ec2.New(sess, &aws.Config{
			Credentials: creds,
			Region:      awsRegion,
		})

		// EC2 instances
		match, _ = regexp.MatchString("i-[a-zA-Z0-9_]+", id)
		if match {
			items := describeEc2(ec2Svc, []*string{awsID}, accountID, accountAlias, verbose)
			itemsChannel <- items
		}

		// Security Groups
		match, _ = regexp.MatchString("sg-[a-zA-Z0-9_]+", id)
		if match {
			items := describeSg(ec2Svc, []*string{awsID}, accountID, accountAlias, verbose)
			itemsChannel <- items
		}

	}
}

func describeSg(ec2Svc *ec2.EC2, IDs []*string, account, accountAlias string, verbose bool) Items {
	var items []Item

	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: IDs,
	}

	// Call to get detailed information on each instance
	result, err := ec2Svc.DescribeSecurityGroups(input)
	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if verbose || !match {
			log.Println("Error", err)
		}
		return items
	}

	for _, sg := range result.SecurityGroups {

		var tags []*ec2.Tag
		for _, tag := range sg.Tags {
			tags = append(tags, tag)
		}

		item := Item{
			ID:   *sg.GroupId,
			Type: "Security Group",
			Tags: tags,
			Account: account,
			AccountAlias: accountAlias,
		}

		items = append(items, item)
	}

	return items
}

func describeEc2(ec2Svc *ec2.EC2, IDs []*string, account, accountAlias string, verbose bool) Items {
	var items []Item

	input := &ec2.DescribeInstancesInput{
		InstanceIds: IDs,
	}

	// Call to get detailed information on each instance
	result, err := ec2Svc.DescribeInstances(input)

	if verbose {
		log.Println(IDs)
		log.Println(input)
		log.Println(result)
	}

	if err != nil {
		match, _ := regexp.MatchString("does not exist", err.Error())
		if verbose || !match {
			log.Println("Error", err)
		}
		return items
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {

			var tags []*ec2.Tag
			for _, tag := range instance.Tags {
				tags = append(tags, tag)
			}

			item := Item{
				ID:   *instance.InstanceId,
				Type: "EC2 instance",
				Tags: tags,
				Account: account,
				AccountAlias: accountAlias,
			}

			items = append(items, item)
		}
	}

	return items
}
