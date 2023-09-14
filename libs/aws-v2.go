package libs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func newAWSV2Config(awsConfig AWSConfig, region string) (aws.Config, error) {
	var cfg aws.Config
	var err error

	if awsConfig.Type == "iam" {
		client := sts.NewFromConfig(cfg)

		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(aws.NewCredentialsCache(
				stscreds.NewAssumeRoleProvider(
					client,
					awsConfig.IAMRoleARN,
				)),
			),
		)
	} else if awsConfig.Type == "profile" {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithSharedConfigProfile(awsConfig.Profile),
		)
	}

	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func getAccountId(cfg aws.Config) (string, error) {
	client := sts.NewFromConfig(cfg)
	input := &sts.GetCallerIdentityInput{}

	req, err := client.GetCallerIdentity(context.TODO(), input)
	if err != nil {
		return "", err
	}

	return *req.Account, nil
}
