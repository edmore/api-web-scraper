package resource_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/edmore/api-web-scraper/resource"
	"github.com/edmore/api-web-scraper/service"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/redisstorage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource", func() {
	Describe("GetPageContents", func() {
		It("returns a HTTP 200 with correct response for valid inputs", func() {
			// configure defaultCollector
			defaultCollector := colly.NewCollector(
				colly.MaxDepth(1),
				colly.Async(true),
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

			// add storage to the collector
			err := defaultCollector.SetStorage(storage)
			if err != nil {
				panic(err)
			}

			expectedResponse := []byte(`
			{
				"results": {
					"htmlVersion": "HTML 5",
					"title": "Example Domain",
					"headingsCountByLevel": {
						"h1": 1
					},
					"links": [
						{
							"url": "http://www.iana.org/domains/reserved",
							"statusCode": 200,
							"isInternal": false,
							"isAccessible": true
						}
					],
					"linksCount": {
						"accessible": 1,
						"external": 1
					},
					"hasLoginForm": false
				}
			}`)
			srv := service.NewCollector(defaultCollector, storage)
			request := httptest.NewRequest("GET", "/scraper/page-contents?url=https://www.example.com", nil)
			rw := httptest.NewRecorder()
			router := resource.NewRouter(srv)

			router.ServeHTTP(rw, request)

			resp := rw.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(expectedResponse).To(MatchJSON(string(body)))
		})
	})
})
