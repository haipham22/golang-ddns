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
	log           *zap.SugaredLogger
	cloudflareAPI *cloudflare.API
	requestClient http.APIClient
}

type UpdatedDomainInfo struct {
	Name         string
	ZoneIdentity string
	DNS          []*DomainDNSInfo
}

type DomainDNSInfo struct {
	Name      string
	TTL       int
	UseProxy  bool
	CurrentIP string
}

func NewUpdateDns(logger *zap.SugaredLogger, cloudflareAPI *cloudflare.API, requestClient http.APIClient) *UpdateDns {
	return &UpdateDns{
		log:           logger,
		cloudflareAPI: cloudflareAPI,
		requestClient: requestClient,
	}
}

func (u *UpdateDns) Run() error {
	currentIP, err := u.checkCurrentIP()
	if err != nil {
		return err
	}

	u.log.Infof("current ip address is: %s", currentIP)

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

func (u *UpdateDns) mappingAllDomainInfo(domains []config.Domain) *hashset.Set {
	mapAllDomainConfig := hashset.New()

	listZones, err := u.cloudflareAPI.ListZones()
	if err != nil {
		u.log.Errorf("c.cloudflareAPI.ListZone error when get all zone, detail: %s", err)
		return nil
	}

	for _, domain := range domains {

		var (
			zoneIdentifier string
			dns            []*DomainDNSInfo
		)

		if domain.ZoneIdentifier == "" {
			for _, results := range listZones.Result {
				if domain.Name == results.Name {
					zoneIdentifier = results.Id
					break
				}
			}
		}

		u.cloudflareAPI.SearchForDNSRecord()

		if len(domain.DNS) == 0 {
			for _, dns := range domain.DNS {
			}
		}

		mapAllDomainConfig.Add(UpdatedDomainInfo{
			Name:         domain.Name,
			ZoneIdentity: zoneIdentifier,
			DNS:          dns,
		})
	}

	return mapAllDomainConfig
}

func (u *UpdateDns) checkAndUpdateByDomain(domain config.Domain, ip string) {
	if len(domain.DNS) == 0 {
		u.log.Infof("c.checkAndUpdateByDomain: empty DNS list, skip update...")
		return
	}
	u.log.Infof("start check from zoneID: %s", domain.ZoneIdentifier)

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
	u.log.Infof("c.shouldCreateNewDnsRecord: create new dns %#v for zone %s with ip %s", dns, zone, ip)
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
		u.log.Errorf("c.shouldCreateNewDnsRecord: error when create new dns record %+v, maybe try create manual, detail %s", requestBody, err)
		return
	}
	u.log.Infof("Create new A record: %s, zoneId: %s, ip: %s", dns.Name, zone, ip)
}

func (u *UpdateDns) updateDNSRecord(ip, zone string, dnsRecord *cloudflare.DetailDNSRecordData, dns config.DNS) {
	u.log.Infof("c.updateDNSRecord: update dns %#v for zone %s with ip %s", dns, zone, ip)
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
		u.log.Errorf("c.shouldCreateNewDnsRecord: error when create new dns dnsRecord %+v, maybe try create manual, detail %s", requestBody, err)
		return
	}
	u.log.Infof("Update for A record: %s, zoneId: %s, newIP: %s, oldIP: %s", dns.Name, zone, ip, dnsRecord.Content)
}
