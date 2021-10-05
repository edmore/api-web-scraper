package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/edmore/api-web-scraper/model"
	"github.com/gocolly/colly"
	"github.com/gocolly/redisstorage"
)

type CollectorService interface {
	Visit(url string) error
	GetPageTitle() string
	GetHtmlVersion() string
	GetLinks() ([]model.Link, map[string]int)
	GetHeadings() map[string]int
	HasLoginForm() bool
	Reset() error
}

var doctypes = make(map[string]string)

type Collector struct {
	DefaultCollector  *colly.Collector
	DelegateCollector *colly.Collector
	Storage           *redisstorage.Storage
}

func NewCollector(defaultCollector *colly.Collector, storage *redisstorage.Storage) *Collector {
	delegateCollector := defaultCollector.Clone()
	defaultCollector.MaxDepth = 2

	collector := &Collector{defaultCollector, delegateCollector, storage}

	collector.init()
	return collector
}

func (c *Collector) Visit(url string) error {
	c.setValue("url", url, 0)

	c.DefaultCollector.Visit(url)
	c.DefaultCollector.Wait()

	return nil
}

func (c *Collector) init() error {
	c.setDocTypes()

	// registers callbacks
	c.registerPageTitleCallback()
	c.registerHtmlVersionCallback()
	c.registerFollowLinksCallback()
	c.registerInaccessibleLinksCallback()
	c.registerAccessibleLinksCallback()
	c.registerHtmlCallback()

	return nil
}

func (c *Collector) registerPageTitleCallback() error {
	c.DefaultCollector.OnHTML("title", func(e *colly.HTMLElement) {
		c.setValue("title", e.Text, 0)
	})
	return nil
}

func (c *Collector) registerHtmlVersionCallback() error {
	c.DefaultCollector.OnResponse(func(r *colly.Response) {

		bytesReader := bytes.NewReader(r.Body)
		bufReader := bufio.NewReader(bytesReader)
		doctype, _, _ := bufReader.ReadLine()

		c.setValue("htmlVersion", determineDoctype(string(doctype)), 0)
	})

	return nil
}

func (c *Collector) registerHtmlCallback() error {
	c.DefaultCollector.OnHTML("html", func(html *colly.HTMLElement) {
		var headings = make(map[string]int)
		var numberOfPasswordFields int

		html.ForEach("h1,h2,h3,h4,h5,h6", func(index int, hElement *colly.HTMLElement) {
			headings[hElement.Name]++
		})

		html.ForEach("input[type=password]", func(n int, hElement *colly.HTMLElement) {
			numberOfPasswordFields++
		})

		h, _ := json.Marshal(headings)
		c.setBytes("headings", h, 0)
		c.setIntValue("numberOfPasswordFields", numberOfPasswordFields, 0)

	})

	return nil
}

func (c *Collector) registerFollowLinksCallback() error {

	c.DefaultCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))

		if link != "" {
			c.DelegateCollector.Visit(link)
			c.DelegateCollector.Wait()
		}
	})
	return nil
}

func (c *Collector) registerAccessibleLinksCallback() error {

	c.DelegateCollector.OnResponse(func(r *colly.Response) {
		IsAccessible := true

		parsedPageURL, _ := url.Parse(c.getValue("url"))
		followedURL := r.Request.URL
		isInternal := isInternal(parsedPageURL, followedURL)

		// save all accessibleLinks
		c.addLinks(IsAccessible, isInternal, r.StatusCode, followedURL)
	})

	return nil
}

func (c *Collector) registerInaccessibleLinksCallback() error {

	c.DelegateCollector.OnError(func(r *colly.Response, err error) {
		IsAccessible := false

		parsedPageURL, _ := url.Parse(c.getValue("url"))
		followedURL := r.Request.URL
		isInternal := isInternal(parsedPageURL, followedURL)

		// save all inaccessibleLinks
		c.addLinks(IsAccessible, isInternal, r.StatusCode, followedURL)
	})

	return nil
}

func (c *Collector) addLinks(IsAccessible bool, isInternal bool,
	statusCode int, followedURL *url.URL) {
	link, _ := json.Marshal(model.Link{
		Url:          followedURL.String(),
		StatusCode:   statusCode,
		IsInternal:   isInternal,
		IsAccessible: IsAccessible,
	})

	c.Storage.Client.LPush("links", link)
}

func (c *Collector) Reset() error {
	keys := []string{
		"htmlVersion", "title", "headings",
		"numberOfPasswordFields", "url", "links",
	}

	if err := c.Storage.Clear(); err != nil {
		return err
	}

	return c.Storage.Client.Del(keys...).Err()
}

func (c *Collector) GetPageTitle() string {

	return c.getValue("title")
}

func (c *Collector) GetHtmlVersion() string {

	return c.getValue("htmlVersion")
}

func (c *Collector) GetLinks() ([]model.Link, map[string]int) {
	var links = []model.Link{}
	var link = model.Link{}
	var linkCounts = map[string]int{}
	v := c.Storage.Client.LLen("links").Val()

	var i int64
	for i = 0; i < v; i++ {

		l, _ := c.Storage.Client.LPop("links").Bytes()
		_ = json.Unmarshal(l, &link)

		links = append(links, link)
		if link.IsAccessible {
			linkCounts["accessible"]++
		} else {
			linkCounts["inaccessible"]++
		}

		if link.IsInternal {
			linkCounts["internal"]++
		} else {
			linkCounts["external"]++
		}
	}

	return links, linkCounts
}

func (c *Collector) GetHeadings() map[string]int {
	var result = make(map[string]int)
	p, _ := c.getBytes("headings")

	_ = json.Unmarshal(p, &result)
	return result
}

func (c *Collector) HasLoginForm() bool {
	v, _ := c.getIntValue("numberOfPasswordFields")
	return (v == 1)
}

func (c *Collector) setDocTypes() {
	doctypes["HTML 4.01 Strict"] = `"-//W3C//DTD HTML 4.01//EN"`
	doctypes["HTML 4.01 Transitional"] = `"-//W3C//DTD HTML 4.01 Transitional//EN"`
	doctypes["HTML 4.01 Frameset"] = `"-//W3C//DTD HTML 4.01 Frameset//EN"`
	doctypes["XHTML 1.0 Strict"] = `"-//W3C//DTD XHTML 1.0 Strict//EN"`
	doctypes["XHTML 1.0 Transitional"] = `"-//W3C//DTD XHTML 1.0 Transitional//EN"`
	doctypes["XHTML 1.0 Frameset"] = `"-//W3C//DTD XHTML 1.0 Frameset//EN"`
	doctypes["XHTML 1.1"] = `"-//W3C//DTD XHTML 1.1//EN"`
	doctypes["HTML 5"] = `<!DOCTYPE html>`
}

func determineDoctype(html string) string {
	var version = "COULD_NOT_BE_ESTABLISHED"

	for doctype, matcher := range doctypes {
		match := strings.Contains(
			strings.ToLower(html),
			strings.ToLower(matcher))

		if match == true {
			version = doctype
		}
	}

	return version
}

func isInternal(pageURL *url.URL, followedURL *url.URL) bool {

	if followedURL.IsAbs() && pageURL.Host != followedURL.Host {
		return false
	}

	return true
}

// Storage helpers
func (c *Collector) setValue(key string, val string, expiration time.Duration) {
	c.Storage.Client.Set(key, val, 0)
}

func (c *Collector) setBytes(key string, val []byte, expiration time.Duration) {
	c.Storage.Client.Set(key, val, 0)
}

func (c *Collector) setIntValue(key string, val int, expiration time.Duration) {
	c.Storage.Client.Set(key, val, 0)
}

func (c *Collector) getIntValue(key string) (int, error) {
	return c.Storage.Client.Get(key).Int()
}

func (c *Collector) getValue(key string) string {
	return c.Storage.Client.Get(key).Val()
}

func (c *Collector) getBytes(key string) ([]byte, error) {
	return c.Storage.Client.Get(key).Bytes()
}
