package cloudflare

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/zones"
)

var isCloudflareID = regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString

func LookupZone(client *cloudflare.Client, zoneIdentifier string) (id string, name string, err error) {
	if isCloudflareID(zoneIdentifier) {
		id = zoneIdentifier
		cacheKey := ZoneCacheKey(id)
		if cachedName, found := GetID(cacheKey); found {
			return id, cachedName, nil
		}
		zoneDetails, err := client.Zones.Get(context.Background(), zones.ZoneGetParams{ZoneID: cloudflare.F(id)})
		if err != nil {
			return "", "", err
		}
		name = zoneDetails.Name
	} else {
		name = zoneIdentifier
		cacheKey := ZoneCacheKey(name)
		if cachedID, found := GetID(cacheKey); found {
			return cachedID, name, nil
		}
		zoneList, err := client.Zones.List(context.Background(), zones.ZoneListParams{Name: cloudflare.String(name)})
		if err != nil {
			return "", "", err
		}
		if len(zoneList.Result) == 0 {
			return "", "", fmt.Errorf("zone %q not found", name)
		}
		id = zoneList.Result[0].ID
		name = zoneList.Result[0].Name
	}

	SetID(ZoneCacheKey(name), id)
	SetID(ZoneCacheKey(id), name)
	return id, name, nil
}

func LookupDNSRecord(client *cloudflare.Client, zoneID, zoneName, recordIdentifier string) (id string, name string, err error) {
	if isCloudflareID(recordIdentifier) {
		id = recordIdentifier
		cacheKey := DNSRecordCacheKeyByID(id)
		if cachedName, found := GetID(cacheKey); found {
			return id, cachedName, nil
		}
		recordDetails, err := client.DNS.Records.Get(context.Background(), recordIdentifier, dns.RecordGetParams{ZoneID: cloudflare.F(zoneID)})
		if err != nil {
			return "", "", err
		}
		name = recordDetails.Name
	} else {
		name = recordIdentifier
		if name == "@" || name == zoneName {
			name = zoneName
		} else {
			name = strings.TrimSuffix(name, "."+zoneName) + "." + zoneName
		}
		cacheKey := DNSRecordCacheKey(zoneID, name)
		if cachedID, found := GetID(cacheKey); found {
			return cachedID, name, nil
		}
		recordList, err := client.DNS.Records.List(context.Background(), dns.RecordListParams{
			ZoneID: cloudflare.F(zoneID),
			Name:   cloudflare.F(dns.RecordListParamsName{Exact: cloudflare.F(name)}),
		})
		if err != nil {
			return "", "", err
		}
		if len(recordList.Result) == 0 {
			return "", "", fmt.Errorf("record %q not found in zone %s", name, zoneName)
		}
		id = recordList.Result[0].ID
		name = recordList.Result[0].Name
	}

	SetID(DNSRecordCacheKey(zoneID, name), id)
	SetID(DNSRecordCacheKeyByID(id), name)
	return id, name, nil
}

func LookupDNSRecordID(client *cloudflare.Client, zoneID, zoneName, recordIdentifier string) (string, error) {
	id, _, err := LookupDNSRecord(client, zoneID, zoneName, recordIdentifier)
	return id, err
}
