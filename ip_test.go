package echo

import (
	"net"
	"net/http"
	"strings"
	"testing"

	testify "github.com/stretchr/testify/assert"
)

const (
	// For RemoteAddr
	ipForRemoteAddrLoopback  = "127.0.0.1" // From 127.0.0.0/8
	sampleRemoteAddrLoopback = ipForRemoteAddrLoopback + ":8080"
	ipForRemoteAddrExternal  = "203.0.113.1"
	sampleRemoteAddrExternal = ipForRemoteAddrExternal + ":8080"
	// For x-real-ip
	ipForRealIP = "203.0.113.10"
	// For XFF
	ipForXFF1LinkLocal = "169.254.0.101" // From 169.254.0.0/16
	ipForXFF2Private   = "192.168.0.102" // From 192.168.0.0/16
	ipForXFF3External  = "2001:db8::103"
	ipForXFF4Private   = "fc00::104" // From fc00::/7
	ipForXFF5External  = "198.51.100.105"
	ipForXFF6External  = "192.0.2.106"
	ipForXFFBroken     = "this.is.broken.lol"
	// keys for test cases
	ipTestReqKeyNoHeader             = "no header"
	ipTestReqKeyRealIPExternal       = "x-real-ip; remote addr external"
	ipTestReqKeyRealIPInternal       = "x-real-ip; remote addr internal"
	ipTestReqKeyRealIPAndXFFExternal = "x-real-ip and xff; remote addr external"
	ipTestReqKeyRealIPAndXFFInternal = "x-real-ip and xff; remote addr internal"
	ipTestReqKeyXFFExternal          = "xff; remote addr external"
	ipTestReqKeyXFFInternal          = "xff; remote addr internal"
	ipTestReqKeyBrokenXFF            = "broken xff"
)

var (
	sampleXFF = strings.Join([]string{
		ipForXFF6External, ipForXFF5External, ipForXFF4Private, ipForXFF3External, ipForXFF2Private, ipForXFF1LinkLocal,
	}, ", ")

	requests = map[string]*http.Request{
		ipTestReqKeyNoHeader: &http.Request{
			RemoteAddr: sampleRemoteAddrExternal,
		},
		ipTestReqKeyRealIPExternal: &http.Request{
			Header: http.Header{
				"X-Real-Ip": []string{ipForRealIP},
			},
			RemoteAddr: sampleRemoteAddrExternal,
		},
		ipTestReqKeyRealIPInternal: &http.Request{
			Header: http.Header{
				"X-Real-Ip": []string{ipForRealIP},
			},
			RemoteAddr: sampleRemoteAddrLoopback,
		},
		ipTestReqKeyRealIPAndXFFExternal: &http.Request{
			Header: http.Header{
				"X-Real-Ip":         []string{ipForRealIP},
				HeaderXForwardedFor: []string{sampleXFF},
			},
			RemoteAddr: sampleRemoteAddrExternal,
		},
		ipTestReqKeyRealIPAndXFFInternal: &http.Request{
			Header: http.Header{
				"X-Real-Ip":         []string{ipForRealIP},
				HeaderXForwardedFor: []string{sampleXFF},
			},
			RemoteAddr: sampleRemoteAddrLoopback,
		},
		ipTestReqKeyXFFExternal: &http.Request{
			Header: http.Header{
				HeaderXForwardedFor: []string{sampleXFF},
			},
			RemoteAddr: sampleRemoteAddrExternal,
		},
		ipTestReqKeyXFFInternal: &http.Request{
			Header: http.Header{
				HeaderXForwardedFor: []string{sampleXFF},
			},
			RemoteAddr: sampleRemoteAddrLoopback,
		},
		ipTestReqKeyBrokenXFF: &http.Request{
			Header: http.Header{
				HeaderXForwardedFor: []string{ipForXFFBroken + ", " + ipForXFF1LinkLocal},
			},
			RemoteAddr: sampleRemoteAddrLoopback,
		},
	}
)

func TestExtractIP(t *testing.T) {
	_, ipv4AllRange, _ := net.ParseCIDR("0.0.0.0/0")
	_, ipv6AllRange, _ := net.ParseCIDR("::/0")
	_, ipForXFF3ExternalRange, _ := net.ParseCIDR(ipForXFF3External + "/48")
	_, ipForRemoteAddrExternalRange, _ := net.ParseCIDR(ipForRemoteAddrExternal + "/24")

	tests := map[string]*struct {
		extractor   IPExtractor
		expectedIPs map[string]string
	}{
		"ExtractIPDirect": {
			ExtractIPDirect(),
			map[string]string{
				ipTestReqKeyNoHeader:             ipForRemoteAddrExternal,
				ipTestReqKeyRealIPExternal:       ipForRemoteAddrExternal,
				ipTestReqKeyRealIPInternal:       ipForRemoteAddrLoopback,
				ipTestReqKeyRealIPAndXFFExternal: ipForRemoteAddrExternal,
				ipTestReqKeyRealIPAndXFFInternal: ipForRemoteAddrLoopback,
				ipTestReqKeyXFFExternal:          ipForRemoteAddrExternal,
				ipTestReqKeyXFFInternal:          ipForRemoteAddrLoopback,
				ipTestReqKeyBrokenXFF:            ipForRemoteAddrLoopback,
			},
		},
		"ExtractIPFromRealIPHeader(default)": {
			ExtractIPFromRealIPHeader(),
			map[string]string{
				ipTestReqKeyNoHeader:             ipForRemoteAddrExternal,
				ipTestReqKeyRealIPExternal:       ipForRemoteAddrExternal,
				ipTestReqKeyRealIPInternal:       ipForRealIP,
				ipTestReqKeyRealIPAndXFFExternal: ipForRemoteAddrExternal,
				ipTestReqKeyRealIPAndXFFInternal: ipForRealIP,
				ipTestReqKeyXFFExternal:          ipForRemoteAddrExternal,
				ipTestReqKeyXFFInternal:          ipForRemoteAddrLoopback,
				ipTestReqKeyBrokenXFF:            ipForRemoteAddrLoopback,
			},
		},
		"ExtractIPFromRealIPHeader(trust only direct-facing proxy)": {
			ExtractIPFromRealIPHeader(TrustLoopback(false), TrustLinkLocal(false), TrustPrivateNet(false), TrustIPRange(ipForRemoteAddrExternalRange)),
			map[string]string{
				ipTestReqKeyNoHeader:             ipForRemoteAddrExternal,
				ipTestReqKeyRealIPExternal:       ipForRealIP,
				ipTestReqKeyRealIPInternal:       ipForRemoteAddrLoopback,
				ipTestReqKeyRealIPAndXFFExternal: ipForRealIP,
				ipTestReqKeyRealIPAndXFFInternal: ipForRemoteAddrLoopback,
				ipTestReqKeyXFFExternal:          ipForRemoteAddrExternal,
				ipTestReqKeyXFFInternal:          ipForRemoteAddrLoopback,
				ipTestReqKeyBrokenXFF:            ipForRemoteAddrLoopback,
			},
		},
		"ExtractIPFromRealIPHeader(trust direct-facing proxy)": {
			ExtractIPFromRealIPHeader(TrustIPRange(ipForRemoteAddrExternalRange)),
			map[string]string{
				ipTestReqKeyNoHeader:             ipForRemoteAddrExternal,
				ipTestReqKeyRealIPExternal:       ipForRealIP,
				ipTestReqKeyRealIPInternal:       ipForRealIP,
				ipTestReqKeyRealIPAndXFFExternal: ipForRealIP,
				ipTestReqKeyRealIPAndXFFInternal: ipForRealIP,
				ipTestReqKeyXFFExternal:          ipForRemoteAddrExternal,
				ipTestReqKeyXFFInternal:          ipForRemoteAddrLoopback,
				ipTestReqKeyBrokenXFF:            ipForRemoteAddrLoopback,
			},
		},
		"ExtractIPFromXFFHeader(default)": {
			ExtractIPFromXFFHeader(),
			map[string]string{
				ipTestReqKeyNoHeader:             ipForRemoteAddrExternal,
				ipTestReqKeyRealIPExternal:       ipForRemoteAddrExternal,
				ipTestReqKeyRealIPInternal:       ipForRemoteAddrLoopback,
				ipTestReqKeyRealIPAndXFFExternal: ipForRemoteAddrExternal,
				ipTestReqKeyRealIPAndXFFInternal: ipForXFF3External,
				ipTestReqKeyXFFExternal:          ipForRemoteAddrExternal,
				ipTestReqKeyXFFInternal:          ipForXFF3External,
				ipTestReqKeyBrokenXFF:            ipForRemoteAddrLoopback,
			},
		},
		"ExtractIPFromXFFHeader(trust only direct-facing proxy)": {
			ExtractIPFromXFFHeader(TrustLoopback(false), TrustLinkLocal(false), TrustPrivateNet(false), TrustIPRange(ipForRemoteAddrExternalRange)),
			map[string]string{
				ipTestReqKeyNoHeader:             ipForRemoteAddrExternal,
				ipTestReqKeyRealIPExternal:       ipForRemoteAddrExternal,
				ipTestReqKeyRealIPInternal:       ipForRemoteAddrLoopback,
				ipTestReqKeyRealIPAndXFFExternal: ipForXFF1LinkLocal,
				ipTestReqKeyRealIPAndXFFInternal: ipForRemoteAddrLoopback,
				ipTestReqKeyXFFExternal:          ipForXFF1LinkLocal,
				ipTestReqKeyXFFInternal:          ipForRemoteAddrLoopback,
				ipTestReqKeyBrokenXFF:            ipForRemoteAddrLoopback,
			},
		},
		"ExtractIPFromXFFHeader(trust direct-facing proxy)": {
			ExtractIPFromXFFHeader(TrustIPRange(ipForRemoteAddrExternalRange)),
			map[string]string{
				ipTestReqKeyNoHeader:             ipForRemoteAddrExternal,
				ipTestReqKeyRealIPExternal:       ipForRemoteAddrExternal,
				ipTestReqKeyRealIPInternal:       ipForRemoteAddrLoopback,
				ipTestReqKeyRealIPAndXFFExternal: ipForXFF3External,
				ipTestReqKeyRealIPAndXFFInternal: ipForXFF3External,
				ipTestReqKeyXFFExternal:          ipForXFF3External,
				ipTestReqKeyXFFInternal:          ipForXFF3External,
				ipTestReqKeyBrokenXFF:            ipForRemoteAddrLoopback,
			},
		},
		"ExtractIPFromXFFHeader(trust everything)": {
			// This is similar to legacy behavior, but ignores x-real-ip header.
			ExtractIPFromXFFHeader(TrustIPRange(ipv4AllRange), TrustIPRange(ipv6AllRange)),
			map[string]string{
				ipTestReqKeyNoHeader:             ipForRemoteAddrExternal,
				ipTestReqKeyRealIPExternal:       ipForRemoteAddrExternal,
				ipTestReqKeyRealIPInternal:       ipForRemoteAddrLoopback,
				ipTestReqKeyRealIPAndXFFExternal: ipForXFF6External,
				ipTestReqKeyRealIPAndXFFInternal: ipForXFF6External,
				ipTestReqKeyXFFExternal:          ipForXFF6External,
				ipTestReqKeyXFFInternal:          ipForXFF6External,
				ipTestReqKeyBrokenXFF:            ipForRemoteAddrLoopback,
			},
		},
		"ExtractIPFromXFFHeader(trust ipForXFF3External)": {
			// This trusts private network also after "additional" trust ranges unlike `TrustNProxies(1)` doesn't
			ExtractIPFromXFFHeader(TrustIPRange(ipForXFF3ExternalRange)),
			map[string]string{
				ipTestReqKeyNoHeader:             ipForRemoteAddrExternal,
				ipTestReqKeyRealIPExternal:       ipForRemoteAddrExternal,
				ipTestReqKeyRealIPInternal:       ipForRemoteAddrLoopback,
				ipTestReqKeyRealIPAndXFFExternal: ipForRemoteAddrExternal,
				ipTestReqKeyRealIPAndXFFInternal: ipForXFF5External,
				ipTestReqKeyXFFExternal:          ipForRemoteAddrExternal,
				ipTestReqKeyXFFInternal:          ipForXFF5External,
				ipTestReqKeyBrokenXFF:            ipForRemoteAddrLoopback,
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := testify.New(t)
			for key, req := range requests {
				actual := test.extractor(req)
				expected := test.expectedIPs[key]
				assert.Equal(expected, actual, "Request: %s", key)
			}
		})
	}
}
