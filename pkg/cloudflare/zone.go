package cloudflare

import (
	"encoding/json"
	"fmt"
	"goland-ddns/pkg/http"
	"time"
)

type Results struct {
	Account struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"account"`
	ActivatedOn     time.Time `json:"activated_on"`
	CreatedOn       time.Time `json:"created_on"`
	DevelopmentMode int       `json:"development_mode"`
	Id              string    `json:"id"`
	Meta            struct {
		CdnOnly                bool `json:"cdn_only"`
		CustomCertificateQuota int  `json:"custom_certificate_quota"`
		DnsOnly                bool `json:"dns_only"`
		FoundationDns          bool `json:"foundation_dns"`
		PageRuleQuota          int  `json:"page_rule_quota"`
		PhishingDetected       bool `json:"phishing_detected"`
		Step                   int  `json:"step"`
	} `json:"meta"`
	ModifiedOn          time.Time `json:"modified_on"`
	Name                string    `json:"name"`
	OriginalDnsHost     string    `json:"original_dnshost"`
	OriginalNameServers []string  `json:"original_name_servers"`
	OriginalRegistrar   string    `json:"original_registrar"`
	Owner               struct {
		Id   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"owner"`
	VanityNameServers []string `json:"vanity_name_servers"`
}

func (c *API) ListZones() (*BaseCloudflareResponse[[]*Results], error) {
	bytes, err := c.request.Get(http.RequestParams{
		URL:    fmt.Sprintf("%s/zones", cloudflareBaseEndpoint),
		Header: c.commonHeader(),
	})

	if err != nil {
		return nil, fmt.Errorf("c.ListZones error when get all zone, detail %s", err)
	}

	var response *BaseCloudflareResponse[[]*Results]
	err = json.Unmarshal(bytes, response)
	if err != nil {
		return nil, fmt.Errorf("c.ListZones invalid response data format, detail %s", err)
	}

	return response, nil
}
