package cloudflare

import (
	"encoding/json"
	"fmt"
	"time"

	errorsLib "github.com/pkg/errors"

	"goland-ddns/pkg/http"
)

type DNSRequestBody struct {
	Content string   `json:"content"`
	Name    string   `json:"name"`
	Proxied bool     `json:"proxied"`
	Type    string   `json:"type"`
	Comment string   `json:"comment"`
	Tags    []string `json:"tags"`
	Ttl     int      `json:"ttl"`
}

func (c *API) CreateDNSRecord(zone string, dns DNSRequestBody) (*DetailDNSRecordResponse, error) {
	bytes, err := c.request.Post(http.RequestParams{
		URL:    fmt.Sprintf("%s/zones/%s/dns_records", cloudflareBaseEndpoint, zone),
		Header: c.commonHeader(),
		Body:   dns,
	})
	if err != nil {
		return nil, fmt.Errorf("c.CreateDnsRecord error when create, detail %s", err)
	}

	var response *DetailDNSRecordResponse

	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return nil, fmt.Errorf("c.CreateDnsRecord invalid response data format, detail %s", err)
	}

	return response, nil
}

func (c *API) UpdateDNSRecord(zone string, identifier string, dns DNSRequestBody) (*DetailDNSRecordResponse, error) {
	bytes, err := c.request.Patch(http.RequestParams{
		URL:    fmt.Sprintf("%s/zones/%s/dns_records/%s", cloudflareBaseEndpoint, zone, identifier),
		Header: c.commonHeader(),
		Body:   dns,
	})
	if err != nil {
		return nil, fmt.Errorf("c.UpdateDnsRecord error when update, detail %s", err)
	}

	var response *DetailDNSRecordResponse

	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return nil, fmt.Errorf("c.UpdateDNSRecord invalid response data format, detail %s", err)
	}

	return response, nil
}

// DetailDNSRecordResponse information about a DNS record
type DetailDNSRecordResponse struct {
	Errors   []ErrorResponseData `json:"errors"`
	Result   DetailDNSRecordData `json:"result"`
	Success  bool                `json:"success"`
	Messages []interface{}       `json:"messages"`
}

// ListDNSRecordResponse information about a DNS record
type ListDNSRecordResponse struct {
	Errors   []ErrorResponseData   `json:"errors"`
	Result   []DetailDNSRecordData `json:"result"`
	Success  bool                  `json:"success"`
	Messages []interface{}         `json:"messages"`
}

// DetailDNSRecordData provides detailed information about a DNS record.
type DetailDNSRecordData struct {
	ID        string   `json:"id"`        // An identifier for the DNS record.
	ZoneID    string   `json:"zone_id"`   // Identifies the DNS zone that the record belongs to.
	ZoneName  string   `json:"zone_name"` // The name of the DNS zone.
	Name      string   `json:"name"`      // The name of the DNS record.
	Type      string   `json:"type"`      // The type of the DNS record, such as A, AAAA, CNAME, etc.
	Content   string   `json:"content"`   // The content of the DNS record, which depends on the type of the record.
	Proxiable bool     `json:"proxiable"` // Indicates if the DNS record is proxiable. This is typically true for A and AAAA record types.
	Proxied   bool     `json:"proxied"`   // Indicates if the DNS record is being proxied. This can offer additional features such as Cloudflare's security and acceleration.
	TTL       int      `json:"ttl"`       // The time to live for the DNS record, in seconds.
	Locked    bool     `json:"locked"`    // Indicates if the DNS record is locked and cannot be modified.
	Meta      struct { // Additional metadata about the DNS record.
		AutoAdded           bool   `json:"auto_added"`             // Indicates if the DNS record was automatically added.
		ManagedByApps       bool   `json:"managed_by_apps"`        // Indicates if the DNS record is managed by an application.
		ManagedByArgoTunnel bool   `json:"managed_by_argo_tunnel"` // Indicates if the DNS record is managed by Argo Tunnel.
		Source              string `json:"source"`                 // Indicates the source of the DNS record, possibly the service which created or manages it.
	} `json:"meta"`
	Comment    interface{}   `json:"comment"`     // A comment associated with the DNS record. The specific type can vary, hence it is declared as an 'interface{}'.
	Tags       []interface{} `json:"tags"`        // Tags associated with the DNS record. Each tag's type can vary, hence they are declared as 'interface{}'.
	CreatedOn  time.Time     `json:"created_on"`  // The time the DNS record was created.
	ModifiedOn time.Time     `json:"modified_on"` // The last time the DNS record was modified.
}

// SearchForDNSRecord returns detailed information about a specific DNS record inside a specific zone.
func (c *API) SearchForDNSRecord(zone string, dns string) (*ListDNSRecordResponse, error) {

	bytes, err := c.request.Get(http.RequestParams{
		URL:    fmt.Sprintf("%s/zones/%s/dns_records?name=%s", cloudflareBaseEndpoint, zone, dns),
		Header: c.commonHeader(),
	})
	if err != nil {
		return nil, fmt.Errorf("c.SearchForDnsRecord error when get, detail %s", err)
	}

	var response *ListDNSRecordResponse

	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return nil, fmt.Errorf("c.SearchForDnsRecord invalid response data format, detail %s", err)
	}

	if !response.Success {
		errorResponse := response.Errors
		for _, errRes := range errorResponse {
			return nil, errorsLib.Wrapf(
				ErrNoRouteMatches,
				"c.SearchForDnsRecord request return failed with code: %v, message %s",
				errRes.Code,
				errRes.Message,
			)
		}
	}

	return response, nil
}

func (c *API) AllDnsByZone(zone string) (*ListDNSRecordResponse, error) {

}
