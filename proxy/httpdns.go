package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/miekg/dns"
	"net/url"
	"strconv"
	"strings"
)

const (
	HttpDnsUrlPathPrefix    = "/resolve"
	HttpDnsUrlPathPrefixBak = "/dns-query"
)

var HTTPDNSSupportType = map[string]uint16{
	"":      dns.TypeA,
	"A":     dns.TypeA,
	"AAAA":  dns.TypeAAAA,
	"NS":    dns.TypeNS,
	"PTR":   dns.TypePTR,
	"TXT":   dns.TypeTXT,
	"SOA":   dns.TypeSOA,
	"SPF":   dns.TypeSPF,
	"SRV":   dns.TypeSRV,
	"MX":    dns.TypeMX,
	"NAPTR": dns.TypeNAPTR,
	"CNAME": dns.TypeCNAME,
	"DNAME": dns.TypeDNAME,
	"CAA":   dns.TypeCAA,
}

const (
	HttpDnsAnswerTypeDoh uint8 = iota
	HttpDnsAnswerTypeJsonAnswer
)

type DnsResponse struct {
	TC       bool       `json:"tc"`
	RD       bool       `json:"rd"`
	RA       bool       `json:"ra"`
	AD       bool       `json:"ad"`
	CD       bool       `json:"cd"`
	Status   int        `json:"status"`
	Question []Question `json:"question"`
	Answer   []RR       `json:"answer"`
}
type Question struct {
	Name string `json:"name"`
	Type uint16 `json:"type"`
}
type RR struct {
	Name string `json:"name"`
	TTL  int    `json:"ttl"`
	Type int    `json:"type"`
	Data string `json:"data"`
}

func parseHTTPArgs(args url.Values) ([]byte, string, error) {
	domainName := convertFQDN(args.Get("name"))
	qTypeString := strings.ToUpper(args.Get("type"))
	remoteHostStr := args.Get("ip")
	qType, ok := HTTPDNSSupportType[qTypeString]
	if !ok {
		return nil, "", fmt.Errorf("msg.Unpack: type %s is invalid", qTypeString)
	}

	if !strings.HasSuffix(domainName, ".") {
		domainName = domainName + "."
	}

	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id: dns.Id(), RecursionDesired: true,
		},
		Question: []dns.Question{
			{Name: domainName, Qtype: qType, Qclass: dns.ClassINET},
		},
	}
	buf, err := msg.Pack()
	return buf, remoteHostStr, err
}

func formatHTTPDNSMsg(msg *dns.Msg, answerType uint8) ([]byte, error) {
	if answerType == HttpDnsAnswerTypeDoh {
		return msg.Pack()
	}
	if len(msg.Question) == 0 {
		return json.Marshal(&DnsResponse{})
	}
	qname := msg.Question[0].Name
	qType := dns.TypeToString[msg.Question[0].Qtype]
	qt := Question{
		Name: qname,
		Type: msg.Question[0].Qtype,
	}
	resp := &DnsResponse{
		Status:   msg.Rcode,
		TC:       msg.Truncated,
		RD:       msg.RecursionDesired,
		RA:       msg.RecursionAvailable,
		AD:       msg.AuthenticatedData,
		CD:       msg.CheckingDisabled,
		Question: []Question{qt},
		Answer:   make([]RR, 0),
	}
	rrs := rrsToArray(msg.Answer)
	for _, rr := range rrs {
		if qType == rr[3] {
			ttl, _ := strconv.Atoi(rr[1])
			r := RR{
				Name: rr[0],
				Type: int(HTTPDNSSupportType[strings.ToUpper(qType)]),
				TTL:  ttl,
				Data: rr[4],
			}
			resp.Answer = append(resp.Answer, r)
		}
	}
	return json.Marshal(&resp)
}

func rrsToArray(rrs []dns.RR) (ret [][]string) {
	for _, rr := range rrs {
		if rr == nil {
			continue
		}
		arr := strings.SplitN(rr.String(), "\t", 5)
		if len(arr) == 5 {
			arr[3] = strings.ToUpper(arr[3])
			ret = append(ret, arr)
		}
	}
	return ret
}
func convertFQDN(domain string) string {
	return strings.TrimSpace(strings.Trim(domain, ".")) + "."
}
