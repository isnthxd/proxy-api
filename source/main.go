package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"proxy-api/db"
	"proxy-api/tasks"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

var allowedProxyTypes = map[string]struct{}{
	"http": {}, "https": {}, "socks4": {}, "socks5": {}, "all": {},
}

func RemoveDuplicates(strings []string) []string {
	unique := make([]string, 0, len(strings))
	seen := make(map[string]struct{})

	for _, s := range strings {
		if _, found := seen[s]; !found {
			seen[s] = struct{}{}
			unique = append(unique, s)
		}
	}

	return unique
}

func getProxiesHandler(database *db.Database) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		proxyType := ctx.Query("type")
		timeoutParam := ctx.Query("timeout")

		if _, ok := allowedProxyTypes[proxyType]; !ok {
			ctx.String(400, "Invalid type parameter. Allowed values are: http, https, socks4, socks5, all.")
			return
		}

		var proxies []string
		if timeoutParam != "" {
			timeout, err := strconv.Atoi(timeoutParam)
			if err == nil {
				proxies = RemoveDuplicates(database.GetOnlineProxies(proxyType, timeout))
			} else {
				ctx.String(400, "Invalid timeout parameter.")
				return
			}
		} else {
			proxies = RemoveDuplicates(database.GetOnlineProxies(proxyType, -1))
		}

		if len(proxies) == 0 {
			ctx.String(400, "No proxies found for the specified type or timeout.")
			return
		}

		ctx.String(200, strings.Join(proxies, "\r\n"))
	}
}

func initializeDatabase() (*db.Database, error) {
	database := &db.Database{}
	if err := database.Init("cache/cache.db"); err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}
	return database, nil
}

func startCronJobs(database *db.Database) {
	c := cron.New()
	c.AddFunc("@every 1h", func() {
		tasks.ScraperTask(database)
		tasks.CheckerTask(database)
	})
	c.Start()
}

func startHttpServer(database *db.Database) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.GET("/proxies", getProxiesHandler(database))
	if err := router.Run(":3000"); err != nil {
		log.Fatalf("Failed to run HTTP server: %v", err)
	}
}

func main() {
	database, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	fmt.Println("Scraping proxies...")
	tasks.ScraperTask(database)
	go tasks.CheckerTask(database)

	fmt.Println("Starting cron jobs...")
	startCronJobs(database)

	fmt.Println("Starting HTTP server on :3000...")
	startHttpServer(database)
}
