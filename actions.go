package main

type Service struct {
	ServiceName    string `json:"serviceName"`
	DetectorID     string `json:"detectorId"`
	Action         interface{} `json:"action"`
	ResourceRole   string `json:"resourceRole"`
	AdditionalInfo struct {
		ThreatListName  string `json:"threatListName"`
		ThreatName         int    `json:"threatName"`
	} `json:"additionalInfo"`
	EventFirstSeen string `json:"eventFirstSeen"`
	EventLastSeen  string `json:"eventLastSeen"`
	Archived       bool   `json:"archived"`
	Count          int    `json:"count"`
}

type PortProbeAction struct {
	ActionType      string `json:"actionType"`
	PortProbeAction struct {
		PortProbeDetails []struct {
			LocalPortDetails struct {
				Port     int    `json:"port"`
				PortName string `json:"portName"`
			} `json:"localPortDetails"`
			RemoteIPDetails struct {
				Country struct {
					CountryName string `json:"countryName"`
				} `json:"country"`
				City struct {
					CityName string `json:"cityName"`
				} `json:"city"`
				GeoLocation struct {
					Lon float64 `json:"lon"`
					Lat float64 `json:"lat"`
				} `json:"geoLocation"`
				Organization struct {
					AsnOrg string `json:"asnOrg"`
					Org    string `json:"org"`
					Isp    string `json:"isp"`
					Asn    string `json:"asn"`
				} `json:"organization"`
				IPAddressV4 string `json:"ipAddressV4"`
			} `json:"remoteIpDetails"`
		} `json:"portProbeDetails"`
		Blocked bool `json:"blocked"`
	} `json:"dnsRequestAction"`
}

type DnsRequestAction struct {
	ActionType      string `json:"actionType"`
	DnsRequestAction struct{
		Domain string `json:"domain"`
		Protocol string  `json:"protocol"`
		Blocked bool `json:"blocked"`
	} `json:"networkConnectionAction"`
}

type NetworkConnectionAction struct {
	ActionType              string `json:"actionType"`
	NetworkConnectionAction struct {
		ConnectionDirection string `json:"connectionDirection"`
		RemoteIPDetails struct {
			IPAddressV4 string `json:"ipAddressV4"`
			Organization struct {
				Asn int    `json:"asn"`
				AsnOrg int    `json:"asnOrg"`
				Isp string `json:"isp"`
				Org string `json:"org"`
			} `json:"organization"`
			Country struct {
				CountryName string `json:"countryName"`
				CountryCode string `json:"countryCode"`
			} `json:"country"`
			City struct {
				CityName string `json:"cityName"`
			} `json:"city"`
			GeoLocation struct {
				Lat int `json:"lat"`
				Lon int `json:"lon"`
			} `json:"geoLocation"`

		} `json:"remoteIpDetails"`
		RemotePortDetails struct {
			Port     int    `json:"port"`
			PortName string `json:"portName"`
		} `json:"remotePortDetails"`
		LocalPortDetails    struct {
			Port     int    `json:"port"`
			PortName string `json:"portName"`
		} `json:"localPortDetails"`
		Protocol        string `json:"protocol"`
		Blocked             bool   `json:"blocked"`
	} `json:"networkConnectionAction"`
}