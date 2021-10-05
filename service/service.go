package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/edmore/api-web-scraper/model"
	"github.com/gocolly/colly"
	"github.com/gocolly/redisstorage"
)

type CollectorService interface {
	// Init() error
	Visit(url string) error
	GetPageTitle() string
	GetHtmlVersion() string
	GetLinksCount() map[string]int
	GetLinks() []model.Link
	GetHeadings() map[string]int
	HasLoginForm() bool
	Reset() error
}

var doctypes = make(map[string]string)
var linksCount = make(map[string]int)

// var links = []model.Link{}

type Collector struct {
	DefaultCollector  *colly.Collector
	DelegateCollector *colly.Collector
	Storage           *redisstorage.Storage
}

func NewCollector(defaultCollector *colly.Collector, storage *redisstorage.Storage) *Collector {
	delegateCollector := defaultCollector.Clone()
	defaultCollector.MaxDepth = 2

	coll := &Collector{defaultCollector, delegateCollector, storage}

	coll.init()
	return coll
}

func (s *Collector) Visit(u string) error {

	s.Storage.Client.Set(
		"url", u, 0)

	s.DefaultCollector.Visit(u)
	s.DefaultCollector.Wait()

	return nil
}

func (s *Collector) init() error {

	s.setDocTypes()
	s.registerPageTitleCallback()
	s.registerHtmlVersionCallback()
	s.registerFollowLinksCallback()
	s.registerInaccessibleLinksCallback()
	s.registerAccessibleLinksCallback()
	s.registerHtmlCallback()
	s.registerOnScrapedCallback()

	return nil
}

func (s *Collector) registerPageTitleCallback() error {
	s.DefaultCollector.OnHTML("title", func(e *colly.HTMLElement) {
		s.Storage.Client.Set("title", e.Text, 0)
	})
	return nil
}

func (s *Collector) registerHtmlVersionCallback() error {
	s.DefaultCollector.OnResponse(func(r *colly.Response) {

		bytesReader := bytes.NewReader(r.Body)
		bufReader := bufio.NewReader(bytesReader)
		doctype, _, _ := bufReader.ReadLine()

		s.Storage.Client.Set(
			"htmlVersion", determineDoctype(string(doctype)), 0)
	})

	return nil
}

func (s *Collector) registerHtmlCallback() error {
	s.DefaultCollector.OnHTML("html", func(html *colly.HTMLElement) {
		var headings = make(map[string]int)
		var numberOfPasswordFields int

		html.ForEach("h1,h2,h3,h4,h5,h6", func(index int, hElement *colly.HTMLElement) {
			headings[hElement.Name]++
		})

		html.ForEach("input[type=password]", func(n int, hElement *colly.HTMLElement) {
			numberOfPasswordFields++
		})

		h, _ := json.Marshal(headings)
		s.Storage.Client.Set(
			"headings", h, 0)

		s.Storage.Client.Set(
			"numberOfPasswordFields", numberOfPasswordFields, 0)

	})

	return nil
}

func (s *Collector) registerFollowLinksCallback() error {

	s.DefaultCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))

		if link != "" {
			s.DelegateCollector.Visit(link)
			s.DelegateCollector.Wait()
		}
	})
	return nil
}

func (s *Collector) registerInaccessibleLinksCallback() error {

	s.DelegateCollector.OnError(func(r *colly.Response, err error) {
		var isInternal bool

		u := s.Storage.Client.Get("url").Val()
		parsedURL, _ := url.Parse(u)
		linksCount["inaccessible"]++

		if r.Request.URL.IsAbs() && parsedURL.Host != r.Request.URL.Host {
			linksCount["external"]++

		} else {
			linksCount["internal"]++
			isInternal = true
		}

		// save all inaccessibleLinks
		l, _ := json.Marshal(model.Link{
			Url:          r.Request.URL.String(),
			StatusCode:   r.StatusCode,
			IsInternal:   isInternal,
			IsAccessible: false,
		})

		s.Storage.Client.LPush("links", l)
	})

	return nil
}

func (s *Collector) registerAccessibleLinksCallback() error {

	s.DelegateCollector.OnResponse(func(r *colly.Response) {
		var isInternal bool
		u := s.Storage.Client.Get("url").Val()
		parsedURL, _ := url.Parse(u)
		linksCount["accessible"]++

		if r.Request.URL.IsAbs() && parsedURL.Host != r.Request.URL.Host {
			linksCount["external"]++
		} else {
			linksCount["internal"]++
			isInternal = true
		}

		// save all accessibleLinks
		// links = append(links, model.Link{
		// 	Url:          r.Request.URL.String(),
		// 	StatusCode:   r.StatusCode,
		// 	IsInternal:   isInternal,
		// 	IsAccessible: true,
		// })

		// save all inaccessibleLinks
		l, _ := json.Marshal(model.Link{
			Url:          r.Request.URL.String(),
			StatusCode:   r.StatusCode,
			IsInternal:   isInternal,
			IsAccessible: true,
		})

		s.Storage.Client.LPush("links", l)
	})

	return nil
}

func (s *Collector) registerOnScrapedCallback() error {
	s.DefaultCollector.OnScraped(func(r *colly.Response) {

		p, _ := json.Marshal(linksCount)
		s.Storage.Client.Set("linksCount", p, 0)

		// q, _ := json.Marshal(links)
		// s.Storage.Client.Set("links", q, 0)

	})

	return nil
}

func (s *Collector) Reset() error {
	// keys := []string{
	// 	"htmlVersion", "title", "linksCount", "headings",
	// 	"numberOfPasswordFields", "url", "links",
	// }

	if err := s.Storage.Clear(); err != nil {
		return err
	}

	return s.Storage.Client.FlushAll().Err()
	// return s.Storage.Client.Del(keys...).Err()
}

func (s *Collector) GetPageTitle() string {

	return s.Storage.Client.Get("title").Val()
}

func (s *Collector) GetHtmlVersion() string {

	return s.Storage.Client.Get("htmlVersion").Val()
}

func (s *Collector) GetLinksCount() map[string]int {
	var result = make(map[string]int)
	p, _ := s.Storage.Client.Get("linksCount").Bytes()

	_ = json.Unmarshal(p, &result)
	return result
}

func (s *Collector) GetLinks() []model.Link {
	var links = []model.Link{}
	var link = model.Link{}
	v := s.Storage.Client.LLen("links").Val()

	// vals, _ := s.Storage.Client.LRange("links", 0, -1).Result()
	// for i, v := range vals {
	var i int64
	for i = 0; i < v; i++ {

		l, _ := s.Storage.Client.LPop("links").Bytes()
		_ = json.Unmarshal(l, &link)

		println(string(l))

		links = append(links, link)
	}

	return links
}

func (s *Collector) GetHeadings() map[string]int {
	var result = make(map[string]int)
	p, _ := s.Storage.Client.Get("headings").Bytes()

	_ = json.Unmarshal(p, &result)
	return result
}

func (s *Collector) HasLoginForm() bool {
	v, _ := s.Storage.Client.Get("numberOfPasswordFields").Int()
	return (v == 1)
}

func (s *Collector) setDocTypes() {
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
	var version = "UNKNOWN"

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
