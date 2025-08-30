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

var isCloudflareID = regexp.MustCompile(`^[a-f09]{32}$`).MatchString

func LookupZone(client *cloudflare.Client, zoneIdentifier string) (id string, name string, err error) {
	if isCloudflareID(zoneIdentifier) {
		id = zoneIdentifier
		cachedName, _ := GetID(ZoneCacheKey(id))
		return id, cachedName, nil
	}

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

	SetID(ZoneCacheKey(name), id)
	SetID(ZoneCacheKey(id), name)

	return id, name, nil
}

func LookupDNSRecordID(client *cloudflare.Client, zoneID, zoneName, recordIdentifier string) (string, error) {
	if isCloudflareID(recordIdentifier) {
		return recordIdentifier, nil
	}

	var fqdn string
	if recordIdentifier == "@" || recordIdentifier == zoneName {
		fqdn = zoneName
	} else {
		fqdn = strings.TrimSuffix(recordIdentifier, "."+zoneName) + "." + zoneName
	}

	cacheKey := DNSRecordCacheKey(zoneID, fqdn)
	if recordID, found := GetID(cacheKey); found {
		return recordID, nil
	}

	recordList, err := client.DNS.Records.List(context.Background(), dns.RecordListParams{
		ZoneID: cloudflare.F(zoneID),
		Name:   cloudflare.F(dns.RecordListParamsName{Exact: cloudflare.F(fqdn)}),
	})
	if err != nil {
		return "", err
	}
	if len(recordList.Result) == 0 {
		return "", fmt.Errorf("record %q not found in zone %s", fqdn, zoneName)
	}

	recordID := recordList.Result[0].ID

	SetID(cacheKey, recordID)

	return recordID, nil
}
