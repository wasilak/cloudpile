package resources

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/wasilak/cloudpile/cache"
)

type Item struct {
	ID             string     `json:"id"`
	Type           string     `json:"type"`
	Tags           []*ec2.Tag `json:"tags"`
	Account        string     `json:"account"`
	AccountAlias   string     `json:"accountAlias"`
	Region         string     `json:"region"`
	IP             string     `json:"ip"`
	PrivateDNSName string     `json:"private_dns_name"`
}

type AWSResourceType interface {
	Init(cache cache.Cache) error
	Get() ([]Item, error)
}

type BaseAWSResource struct {
	Cache        cache.Cache
	Items        []Item
	AccountID    string
	AccountAlias string
}
