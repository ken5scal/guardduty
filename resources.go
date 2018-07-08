package main

type InstanceResource struct {
	ResourceType    string `json:"resourceType"`
	InstanceDetails struct {
		AvailabilityZone  string `json:"availabilityZone"`
		ImageDescription  string `json:"imageDescription"`
		ImageID           string `json:"imageId"`
		InstanceID        string `json:"instanceId"`
		InstanceState     string `json:"instanceState"`
		InstanceType      string `json:"instanceType"`
		LaunchTime        int    `json:"launchTime"`
		NetworkInterfaces []struct {
			Ipv6Addresses      []interface{} `json:"ipv6Addresses"`
			PrivateDNSName     string        `json:"privateDnsName"`
			PrivateIPAddress   string        `json:"privateIpAddress"`
			PrivateIPAddresses []struct {
				PrivateDNSName   string `json:"privateDnsName"`
				PrivateIPAddress string `json:"privateIpAddress"`
			} `json:"privateIpAddresses"`
			PublicDNSName  string `json:"publicDnsName"`
			PublicIP       string `json:"publicIp"`
			SecurityGroups []struct {
				GroupID   string `json:"groupId"`
				GroupName string `json:"groupName"`
			} `json:"securityGroups"`
			SubnetID string `json:"subnetId"`
			VpcID    string `json:"vpcId"`
		} `json:"networkInterfaces"`
		ProductCodes []interface{} `json:"productCodes"`
		Tags         []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"tags"`
	} `json:"instanceDetails"`
}