package tasks

import (
	"net/http"
	"net/url"
	"proxy-api/db"
	"sync"
	"time"

	"github.com/shettyh/threadpool"
	"h12.io/socks"
)

type ProxyCheckTask struct {
	proxy    db.Proxy
	database *db.Database
	wg       *sync.WaitGroup
}

func (t *ProxyCheckTask) Run() {
	defer t.wg.Done()
	status := "offline"
	responseTime := 0

	start := time.Now()
	if checkProxy(t.proxy) {
		status = "online"
		responseTime = int(time.Since(start).Milliseconds())
	}
	t.database.UpdateProxyStatus(t.proxy.Type, t.proxy.Proxy, status, responseTime)
}

func checkProxy(proxy db.Proxy) bool {
	var client *http.Client
	testURL := "http://httpbin.org/ip"

	proxyURL, err := url.Parse("http://" + proxy.Proxy)
	if err != nil {
		return false
	}

	transport := &http.Transport{
		DisableKeepAlives: true,
	}

	switch proxy.Type {
	case "http":
		transport.Proxy = http.ProxyURL(proxyURL)
	case "https":
		testURL = "https://httpbin.org/ip"
		transport.Proxy = http.ProxyURL(proxyURL)
	case "socks4":
		transport.Dial = socks.DialSocksProxy(socks.SOCKS4, proxy.Proxy)
	case "socks5":
		transport.Dial = socks.DialSocksProxy(socks.SOCKS5, proxy.Proxy)
	default:
		return false
	}

	client = &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	resp, err := client.Get(testURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func CheckerTask(database *db.Database) {
	proxies := database.GetAllProxies()
	pool := threadpool.NewThreadPool(1024, int64(len(proxies)))
	var wg sync.WaitGroup

	for _, proxy := range proxies {
		wg.Add(1)
		task := &ProxyCheckTask{
			proxy:    proxy,
			database: database,
			wg:       &wg,
		}
		pool.Execute(task)
	}

	wg.Wait()
	pool.Close()
}
