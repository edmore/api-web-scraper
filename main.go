package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/edmore/api-web-scraper/resource"
	"github.com/edmore/api-web-scraper/service"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/redisstorage"
)

func main() {
	asciiArt := `
	.__                        ___.                                                             
	_____  ______ |__|         __  _  __ ____\_ |__             ______ ________________  ______   ___________ 
	\__  \ \____ \|  |  ______ \ \/ \/ // __ \| __ \   ______  /  ___// ___\_  __ \__  \ \____ \_/ __ \_  __ \
	 / __ \|  |_> >  | /_____/  \     /\  ___/| \_\ \ /_____/  \___ \\  \___|  | \// __ \|  |_> >  ___/|  | \/
	(____  /   __/|__|           \/\_/  \___  >___  /         /____  >\___  >__|  (____  /   __/ \___  >__|   
		 \/|__|                             \/    \/               \/     \/           \/|__|        \/       											
   `

	fmt.Println(asciiArt)
	gin.SetMode(gin.ReleaseMode)

	// configure defaultCollector
	defaultCollector := colly.NewCollector(
		// MaxDepth is 1, so only the links on the scraped page
		// are visited, and no further links are followed
		colly.MaxDepth(1),
		colly.Async(true),
		// colly.Debugger(&debug.LogDebugger{}),
	)
	extensions.RandomUserAgent(defaultCollector)
	defaultCollector.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})
	defaultCollector.Limit(&colly.LimitRule{RandomDelay: 2 * time.Second,
		DomainGlob:  "*",
		Parallelism: 2})

	// create the redis storage
	storage := &redisstorage.Storage{
		Address:  "127.0.0.1:6379",
		Password: "",
		DB:       0,
		Prefix:   "api-web-scraper",
	}

	// add redis storage to the collector
	err := defaultCollector.SetStorage(storage)
	if err != nil {
		panic(err)
	}

	srv := service.NewCollector(defaultCollector, storage)

	router := resource.NewRouter(srv)
	router.Run()
}
