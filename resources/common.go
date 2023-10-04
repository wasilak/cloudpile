package resources

type ItemTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Item struct {
	ID             string    `json:"id"`
	ARN            string    `json:"arn"`
	Type           string    `json:"type"`
	Tags           []ItemTag `json:"tags"`
	Account        string    `json:"account"`
	AccountAlias   string    `json:"accountAlias"`
	Region         string    `json:"region"`
	IP             string    `json:"ip"`
	PrivateDNSName string    `json:"private_dns_name"`
}

type AWSResourceType interface {
	Get() ([]Item, error)
	GetCacheKey() string
}

type BaseAWSResource struct {
	Items        []Item
	AccountID    string
	AccountAlias string
	Region       string
	Type         string
}
