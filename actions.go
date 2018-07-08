package main

type DnsRequestAction struct {
	Action      struct{
		ActionType      string `json:"actionType"`
		DnsRequestAction struct{
			Domain string `json:"domain"`
			Protocol string  `json:"protocol"`
			Blocked bool `json:"blocked"`
		}
	}
}

type PortProbeAction struct {
	Action      struct {
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
		} `json:"portProbeAction"`
	} `json:"action"`
	AdditionalInfo struct {
		ThreatListName  string `json:"threatListName"`
		ThreatName         int    `json:"threatName"`
	} `json:"additionalInfo"`
	Archived       bool   `json:"archived"`
	Count          int    `json:"count"`
	DetectorID     string `json:"detectorId"`
	EventFirstSeen string `json:"eventFirstSeen"`
	EventLastSeen  string `json:"eventLastSeen"`
	ResourceRole   string `json:"resourceRole"`
	ServiceName    string `json:"serviceName"`
}

type NetworkConnectionAction struct {
	Action struct {
		ActionType              string `json:"actionType"`
		NetworkConnectionAction struct {
			Blocked             bool   `json:"blocked"`
			ConnectionDirection string `json:"connectionDirection"`
			LocalPortDetails    struct {
				Port     int    `json:"port"`
				PortName string `json:"portName"`
			} `json:"localPortDetails"`
			Protocol        string `json:"protocol"`
			RemoteIPDetails struct {
				City struct {
					CityName string `json:"cityName"`
				} `json:"city"`
				Country struct {
					CountryName string `json:"countryName"`
				} `json:"country"`
				GeoLocation struct {
					Lat int `json:"lat"`
					Lon int `json:"lon"`
				} `json:"geoLocation"`
				IPAddressV4 string `json:"ipAddressV4"`

				Organization struct {
					Asn int    `json:"asn"`
					Isp string `json:"isp"`
					Org string `json:"org"`
				} `json:"organization"`
			} `json:"remoteIpDetails"`
			RemotePortDetails struct {
				Port     int    `json:"port"`
				PortName string `json:"portName"`
			} `json:"remotePortDetails"`
		} `json:"networkConnectionAction"`
	} `json:"action"`
	AdditionalInfo struct {
		ThreatListName  string `json:"threatListName"`
		Unusual         int    `json:"unusual"`
		UnusualProtocol string `json:"unusualProtocol"`
	} `json:"additionalInfo"`
	Archived       bool   `json:"archived"`
	Count          int    `json:"count"`
	DetectorID     string `json:"detectorId"`
	EventFirstSeen string `json:"eventFirstSeen"`
	EventLastSeen  string `json:"eventLastSeen"`
	ResourceRole   string `json:"resourceRole"`
	ServiceName    string `json:"serviceName"`
}