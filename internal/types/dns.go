package types

import "github.com/cloudflare/cloudflare-go/v6/dns"

type DnsRecordWithZone struct {
	dns.RecordResponse
	ZoneName string
}
