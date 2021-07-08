package getter

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/gammazero/workerpool"

	"github.com/One-Piecs/proxypool/log"

	conf "github.com/One-Piecs/proxypool/config"

	"github.com/One-Piecs/proxypool/pkg/proxy"
	"github.com/One-Piecs/proxypool/pkg/tool"
	"github.com/gocolly/colly"
)

func init() {
	Register("tgchannel", NewTGChannelGetter)
}

type TGChannelGetter struct {
	c         *colly.Collector
	NumNeeded int
	results   []string
	Url       string
	apiUrl    string
}

func NewTGChannelGetter(options tool.Options) (getter Getter, err error) {
	num, found := options["num"]
	t := 200
	switch num.(type) {
	case int:
		t = num.(int)
	case float64:
		t = int(num.(float64))
	}

	if !found || t <= 0 {
		t = 200
	}
	urlInterface, found := options["channel"]
	if found {
		url, err := AssertTypeStringNotNull(urlInterface)
		if err != nil {
			return nil, err
		}
		return &TGChannelGetter{
			c:         tool.GetColly(),
			NumNeeded: t,
			Url:       "https://t.me/s/" + url,
			// apiUrl:    "https://tg.i-c-a.su/rss/" + url,
			apiUrl: conf.Config.TgChannelProxyUrl + url,
		}, nil
	}
	return nil, ErrorUrlNotFound
}

func (g *TGChannelGetter) Get() proxy.ProxyList {
	result := make(proxy.ProxyList, 0)
	g.results = make([]string, 0)
	// 找到所有的文字消息
	g.c.OnHTML("div.tgme_widget_message_text", func(e *colly.HTMLElement) {
		g.results = append(g.results, GrepLinksFromString(e.Text)...)
		// 抓取到http链接，有可能是订阅链接或其他链接，无论如何试一下
		subUrls := urlRe.FindAllString(e.Text, -1)
		for _, url := range subUrls {
			result = append(result, (&Subscribe{Url: url}).Get()...)
		}
	})

	// 找到之前消息页面的链接，加入访问队列
	g.c.OnHTML("link[rel=prev]", func(e *colly.HTMLElement) {
		if len(g.results) < g.NumNeeded {
			_ = e.Request.Visit(e.Attr("href"))
		}
	})

	webStart := time.Now()
	g.results = make([]string, 0)
	err := g.c.Visit(g.Url)
	if err != nil {
		_ = fmt.Errorf("%s", err.Error())
	}

	// 等待并发抓取结果
	g.c.Wait()

	result = append(result, StringArray2ProxyArray(g.results)...)

	log.Infoln("STATISTIC: TGChannel\tcost=%v\tcount=%d\turl=%s\tsub_url=%s",
		time.Since(webStart), len(result), g.Url, "web_message")

	// 获取文件(api需要维护)
	resp, err := tool.GetHttpClient().Get(g.apiUrl)
	if err != nil {
		return result
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	items := strings.Split(string(body), "\n")

	rssStart := time.Now()
	wp := workerpool.New(50)
	m := sync.Mutex{}

	for _, s := range items {
		ss := s
		wp.Submit(func() {
			if strings.Contains(ss, "enclosure url") { // get to xml node
				elements := strings.Split(ss, "\"")
				for _, e := range elements {
					// add 内部部署 http
					if strings.Contains(e, "https://") || strings.Contains(e, "http://") {
						// TODO Webfuzz的可能性比较大，也有可能是订阅链接，为了不拖慢运行速度不写了
						start := time.Now()
						subResult := (&WebFuzz{Url: e}).Get()

						log.Infoln("STATISTIC: TGChannel\tcost=%v\tcount=%d\turl=%s\tsub_url=%s",
							time.Since(start), len(subResult), g.Url, e)
						m.Lock()
						result = append(result, subResult...)
						m.Unlock()
					}
				}
			}
		})
	}
	log.Infoln("STATISTIC: TGChannel\tcost=%v\tcount=%d\turl=%s\tsub_url=%s",
		time.Since(rssStart), len(result), g.Url, "rss_message")
	wp.StopWait()
	return result
}

func (g *TGChannelGetter) Get2ChanWG(pc chan proxy.Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	nodes := g.Get()
	log.Infoln("STATISTIC: TGChannel\tcost=%v\tcount=%d\turl=%s", time.Since(start), len(nodes), g.Url)
	for _, node := range nodes {
		pc <- node
	}
}
func (g *TGChannelGetter) Get2Chan(pc chan proxy.Proxy) {
	nodes := g.Get()
	log.Infoln("STATISTIC: TGChannel\tcount=%d\turl=%s", len(nodes), g.Url)
	for _, node := range nodes {
		pc <- node
	}
}
