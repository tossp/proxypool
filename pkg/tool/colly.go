package tool

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gocolly/colly"
)

func GetColly() *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent(UserAgent),
		colly.MaxDepth(6),
		colly.Async(true),
	)
	c.WithTransport(&http.Transport{
		// Proxy: http.ProxyFromEnvironment,
		Proxy: ProxyURL(),
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, // 超时时间
			KeepAlive: 10 * time.Second, // keepAlive 超时时间
		}).DialContext,
		MaxIdleConns:          100,              // 最大空闲连接数
		IdleConnTimeout:       20 * time.Second, // 空闲连接超时
		TLSHandshakeTimeout:   10 * time.Second, // TLS 握手超时
		ExpectContinueTimeout: 10 * time.Second,
	})
	return c
}

// 配置代理模式
func ProxyURL() func(*http.Request) (*url.URL, error) {
	if getEnvAny("tg_channel_web_proxy") != "" {
		fixedURL, err := url.Parse(getEnvAny("tg_channel_web_proxy"))
		return func(*http.Request) (*url.URL, error) {
			return fixedURL, err
		}
	}
	return http.ProxyFromEnvironment
}

func getEnvAny(names ...string) string {
	for _, n := range names {
		if val := os.Getenv(n); val != "" {
			return val
		}
	}
	return ""
}
