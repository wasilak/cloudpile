package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type LambdaFunction struct {
	Client *lambda.Client
	BaseAWSResource
}

func (r *LambdaFunction) GetCacheKey() string {
	return fmt.Sprintf("%s-%s-%s", r.AccountID, r.Region, r.Type)
}

func (r *LambdaFunction) Get() ([]Item, error) {
	items := []Item{}

	functions, err := r.listFunctions(100)
	if err != nil {
		return nil, err
	}

	for _, function := range functions {

		lambdaTagsInput := lambda.ListTagsInput{
			Resource: function.FunctionArn,
		}

		tagsList, err := r.Client.ListTags(context.TODO(), &lambdaTagsInput)
		if err != nil {
			return nil, err
		}

		tags := []ItemTag{}
		for k, v := range tagsList.Tags {
			newTag := ItemTag{
				Key:   k,
				Value: v,
			}

			tags = append(tags, newTag)
		}

		item := Item{
			Tags:         tags,
			ID:           *function.FunctionName,
			ARN:          *function.FunctionArn,
			Type:         "Lambda function",
			Account:      r.AccountID,
			AccountAlias: r.AccountAlias,
			Region:       r.Region,
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *LambdaFunction) listFunctions(maxItems int) ([]types.FunctionConfiguration, error) {
	var functions []types.FunctionConfiguration
	paginator := lambda.NewListFunctionsPaginator(r.Client, &lambda.ListFunctionsInput{
		MaxItems: aws.Int32(int32(maxItems)),
	})
	for paginator.HasMorePages() && len(functions) < maxItems {
		pageOutput, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		functions = append(functions, pageOutput.Functions...)
	}

	return functions, nil
}
