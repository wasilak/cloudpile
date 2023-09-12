package libs

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
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

func SetupIAMCreds(iamRole string) *credentials.Credentials {
	sess := session.Must(session.NewSession())
	return stscreds.NewCredentials(sess, iamRole)
}

func SetupSharedProfileCreds(sess *session.Session, profile string) *credentials.Credentials {
	return credentials.NewSharedCredentials("", profile)
}

func SetupSession(iamRole, region string, creds *credentials.Credentials) (*session.Session, error) {
	var sess *session.Session

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	})

	if err != nil {
		return sess, err
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return sess, err
	}

	return sess, nil
}
