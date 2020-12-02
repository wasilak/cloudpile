package libs

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func GetIdentity(sess *session.Session) *sts.GetCallerIdentityOutput {
	svc := sts.New(sess)
	input := &sts.GetCallerIdentityInput{}

	result, err := svc.GetCallerIdentity(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Fatal(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Fatal(err.Error())
		}
	}

	return result
}

func getAccountIdFromRoleARN(iamArn string) string {
	re := regexp.MustCompile(`(?m)arn\:aws\:iam\:\:(\w+)\:role.+`)
	return re.ReplaceAllString(iamArn, "$1")
}
