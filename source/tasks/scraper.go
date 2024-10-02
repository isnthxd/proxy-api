package tasks

import (
	"io"
	"net/http"
	"proxy-api/db"
	"regexp"
	"sync"
	"time"

	"github.com/shettyh/threadpool"
)

var proxyPattern = `(?:\b(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?):([1-5][0-9]{4}|[1-9][0-9]{0,4})\b)`

type ScraperCheckTask struct {
	proxySource     db.ProxySource
	database        *db.Database
	existingProxies []db.Proxy
	re              *regexp.Regexp
	wg              *sync.WaitGroup
}

func (t *ScraperCheckTask) Run() {
	defer t.wg.Done()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(t.proxySource.Url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	matches := t.re.FindAllString(string(body), -1)
	for _, match := range matches {
		if !checkIfProxyExist(t.existingProxies, t.proxySource.Type, match) {
			t.database.InsertProxy(t.proxySource.Type, match)
			t.existingProxies = append(t.existingProxies, db.Proxy{Proxy: match, Type: t.proxySource.Type})
		}
	}
}

func checkIfProxyExist(proxies []db.Proxy, proxyType, proxy string) bool {
	for _, p := range proxies {
		if p.Type == proxyType && p.Proxy == proxy {
			return true
		}
	}
	return false
}

func ScraperTask(database *db.Database) {
	proxySources := database.GetAllProxySources()
	existingProxies := database.GetAllProxies()
	re := regexp.MustCompile(proxyPattern)

	pool := threadpool.NewThreadPool(100, int64(len(proxySources)))
	var wg sync.WaitGroup

	for _, source := range proxySources {
		wg.Add(1)
		task := &ScraperCheckTask{
			proxySource:     source,
			database:        database,
			existingProxies: existingProxies,
			re:              re,
			wg:              &wg,
		}
		pool.Execute(task)
	}

	wg.Wait()
	pool.Close()
}
