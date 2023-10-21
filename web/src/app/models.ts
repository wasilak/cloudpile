export interface ItemTag {
    key: string
    value: string
}

export interface AWSResource {
    id: string
    arn: string
    type: string
    tags: Array<ItemTag>
    account: string
    accountAlias: string
    region: string
    ip: string
    private_dns_name: string
}
