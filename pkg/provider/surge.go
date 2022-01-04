package provider

import (
	"strings"

	"github.com/One-Piecs/proxypool/pkg/tool"

	"github.com/One-Piecs/proxypool/pkg/proxy"
)

// Surge provides functions that make proxies support clash client
type Surge struct {
	Base
}

// Provide of Surge generates proxy list supported by surge
func (s Surge) Provide() string {
	s.preFilter()

	var resultBuilder strings.Builder
	for _, p := range *s.Proxies {
		if checkSurgeSupport(p) {
			resultBuilder.WriteString(p.ToSurge() + "\n")
		}
	}
	resultBuilder.WriteString("🐈 ClashX = socks5, 127.0.0.1, 7890" + "\n")
	return resultBuilder.String()
}

func checkSurgeSupport(p proxy.Proxy) bool {
	switch p.(type) {
	case *proxy.ShadowsocksR:
		return false
	case *proxy.Vmess:
		return true
	case *proxy.Shadowsocks:
		ss := p.(*proxy.Shadowsocks)
		if tool.CheckInList(proxy.SSCipherList, ss.Cipher) {
			return true
		}
	case *proxy.Trojan:
		return true
	default:
		return false
	}
	return false
}
