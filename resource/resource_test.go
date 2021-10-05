package resource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

type MockDefaultCollector struct {
	mock.Mock
}

type MockDelegateCollector struct {
	mock.Mock
}

type MockStorage struct {
	mock.Mock
}

func (s *MockStorage) Clear() error {
	args := s.Called()
	return args.Error(1)
}

var _ = Describe("Resource", func() {
	Describe("GetPageContents", func() {
		It("returns a HTTP 200 with correct response for valid inputs", func() {
			// // defaultCollector := new(MockDefaultCollector)
			// // delegateCollector := new(MockDelegateCollector)
			// // storage := MockStorage{}

			// // configure defaultCollector
			// defaultCollector := colly.NewCollector(
			// 	// MaxDepth is 1, so only the links on the scraped page
			// 	// is visited, and no further links are followed
			// 	colly.MaxDepth(1),
			// 	colly.Async(true),
			// 	// colly.Debugger(&debug.LogDebugger{}),
			// )
			// extensions.RandomUserAgent(defaultCollector)
			// // extensions.Referer(collector)
			// defaultCollector.WithTransport(&http.Transport{
			// 	DisableKeepAlives: true,
			// })
			// defaultCollector.Limit(&colly.LimitRule{RandomDelay: 2 * time.Second,
			// 	DomainGlob:  "*",
			// 	Parallelism: 2})

			// // create the redis storage
			// storage := &redisstorage.Storage{
			// 	Address:  "127.0.0.1:6379",
			// 	Password: "",
			// 	DB:       0,
			// 	Prefix:   "api-web-scraper",
			// }

			// // add storage to the collector
			// err := defaultCollector.SetStorage(storage)
			// if err != nil {
			// 	panic(err)
			// }

			// delegateCollector := defaultCollector.Clone()

			// request := httptest.NewRequest("GET", "/scraper/page-contents?url=https://www.example.com", nil)
			// rw := httptest.NewRecorder()
			// router := resource.NewRouter(defaultCollector, delegateCollector, storage)

			// router.ServeHTTP(rw, request)

			// resp := rw.Result()
			// body, _ := ioutil.ReadAll(resp.Body)

			// Expect(resp.StatusCode).To(Equal(http.StatusOK))
			// expectedResponse, _ := json.Marshal("results")
			// Expect(expectedResponse).To(MatchJSON(string(body)))
		})

		It("returns a HTTP 500 if repo error on GetPageContents", func() {

		})
	})
})
