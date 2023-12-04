package update_dns

import (
	"bytes"
	"fmt"
	"net"
	"sync"

	"github.com/emirpasic/gods/sets/hashset"
	errorsLib "github.com/pkg/errors"
	"go.uber.org/zap"

	"goland-ddns/pkg/cloudflare"
	"goland-ddns/pkg/config"
	"goland-ddns/pkg/http"
)

const (
	defaultSeekIPUrl = "https://ipv4.icanhazip.com/"
)

type UpdateDns struct {
	logger        *zap.SugaredLogger
	cloudflareAPI *cloudflare.CloudflareAPI
	requestClient http.APIClient
}

func NewUpdateDns(logger *zap.SugaredLogger, cloudflareAPI *cloudflare.CloudflareAPI, requestClient http.APIClient) *UpdateDns {
	return &UpdateDns{
		logger:        logger,
		cloudflareAPI: cloudflareAPI,
		requestClient: requestClient,
	}
}

func (u *UpdateDns) Run() error {
	currentIP, err := u.checkCurrentIP()
	if err != nil {
		return err
	}

	u.logger.Infof("current ip address is: %s", currentIP)

	var wg sync.WaitGroup

	wg.Add(len(config.ENV.Cloudflare.Domain))

	for _, domain := range config.ENV.Cloudflare.Domain {
		updateDomain := domain
		go func() {
			u.checkAndUpdateByDomain(updateDomain, currentIP)
			defer wg.Done()
		}()
	}
	wg.Wait()
	return nil
}

func (u *UpdateDns) checkCurrentIP() (string, error) {
	seekIPUrl := defaultSeekIPUrl
	if config.ENV.SeekIPURL != "" {
		seekIPUrl = config.ENV.SeekIPURL
	}
	responseByteArr, err := u.requestClient.Get(http.RequestParams{
		URL: seekIPUrl,
	})
	if err != nil {
		return "", fmt.Errorf("error while connect to url, detail %s", err)
	}

	responseByteArr = bytes.Trim(responseByteArr, "\n\r")

	var currentIP = string(responseByteArr)

	if parsedIP := net.ParseIP(currentIP); parsedIP == nil {
		return "", fmt.Errorf("invalid IP format, ip: %s, detail: %s", string(responseByteArr), err.Error())
	}

	return currentIP, nil
}

func (u *UpdateDns) mappingAllDomainInfo(domain []config.Domain) {
	mapAllDomainConfig := hashset.New()
}

func (u *UpdateDns) checkAndUpdateByDomain(domain config.Domain, ip string) {
	if len(domain.DNS) == 0 {
		u.logger.Infof("c.checkAndUpdateByDomain: empty DNS list, skip update...")
		return
	}
	u.logger.Infof("start check from zoneID: %s", domain.ZoneIdentifier)

	for _, dns := range domain.DNS {
		dnsRecord, err := u.seekCurrentDNS(domain.ZoneIdentifier, dns.Name)
		if err != nil {
			u.shouldCreateNewDNSRecord(ip, domain.ZoneIdentifier, dns)
			continue
		}
		u.updateDNSRecord(ip, domain.ZoneIdentifier, dnsRecord, dns)
	}
}

func (u *UpdateDns) seekCurrentDNS(domain string, dns string) (*cloudflare.DetailDNSRecordData, error) {
	dnsRecord, err := u.cloudflareAPI.SearchForDNSRecord(domain, dns)
	if err != nil && errorsLib.Is(err, cloudflare.ErrNoRouteMatches) {
		return nil, fmt.Errorf("no route matchs with dns, detail: %s", err)
	}
	switch len(dnsRecord.Result) {
	case 0:
		return nil, fmt.Errorf("no route matchs with dns, detail: %s", err)
	case 1:
		return &dnsRecord.Result[0], nil
	}
	return nil, nil
}

func (u *UpdateDns) shouldCreateNewDNSRecord(ip, zone string, dns config.DNS) {
	u.logger.Infof("c.shouldCreateNewDnsRecord: create new dns %#v for zone %s with ip %s", dns, zone, ip)
	ttl := dns.TTL

	if ttl <= 0 {
		ttl = 120
	}
	requestBody := cloudflare.DNSRequestBody{
		Content: ip,
		Name:    dns.Name,
		Proxied: dns.UseProxy,
		Type:    "A",
		Comment: "",
		Tags:    nil,
		Ttl:     ttl,
	}
	_, err := u.cloudflareAPI.CreateDNSRecord(zone, requestBody)
	if err != nil {
		u.logger.Errorf("c.shouldCreateNewDnsRecord: error when create new dns record %+v, maybe try create manual, detail %s", requestBody, err)
		return
	}
	u.logger.Infof("Create new A record: %s, zoneId: %s, ip: %s", dns.Name, zone, ip)
}

func (u *UpdateDns) updateDNSRecord(ip, zone string, dnsRecord *cloudflare.DetailDNSRecordData, dns config.DNS) {
	u.logger.Infof("c.updateDNSRecord: update dns %#v for zone %s with ip %s", dns, zone, ip)
	ttl := dns.TTL

	if ttl <= 0 {
		ttl = 120
	}
	requestBody := cloudflare.DNSRequestBody{
		Content: ip,
		Name:    dns.Name,
		Proxied: dns.UseProxy,
		Type:    "A",
		Comment: "",
		Tags:    nil,
		Ttl:     ttl,
	}
	_, err := u.cloudflareAPI.UpdateDNSRecord(zone, dnsRecord.ZoneID, requestBody)
	if err != nil {
		u.logger.Errorf("c.shouldCreateNewDnsRecord: error when create new dns dnsRecord %+v, maybe try create manual, detail %s", requestBody, err)
		return
	}
	u.logger.Infof("Update for A record: %s, zoneId: %s, newIP: %s, oldIP: %s", dns.Name, zone, ip, dnsRecord.Content)
}
