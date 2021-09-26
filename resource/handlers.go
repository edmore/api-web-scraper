package resource

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/edmore/api-web-scraper/model"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/gocolly/redisstorage"
)

type Handler struct {
	defaultCollector  *colly.Collector
	delegateCollector *colly.Collector
	storage           *redisstorage.Storage
}

func (h *Handler) GetPageContents(c *gin.Context) {
	u := c.Query("url")
	// validate URI
	parsedURL, err := url.Parse(u)
	if err != nil {
		log.Fatal("oops")
	}
	// h.storage.Client.Set("url", parsedURL, 0)

	page := model.Page{HeadingsCount: make(map[string]int), PasswordFieldCount: make(map[string]int)}
	links := []model.Link{}
	// headings := []model.Heading{}

	h.defaultCollector.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("url", r.URL.String())
		log.Println("visiting ...", r.URL.String())
	})

	// On every a element which has href attribute call callback
	h.defaultCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		fmt.Printf("link found: %q -> %s\n", e.Text, link)
		if link != "" {
			h.delegateCollector.Visit(link)
			h.delegateCollector.Wait()
		}
	})

	h.defaultCollector.OnHTML("title", func(e *colly.HTMLElement) {
		title := e.Text
		page.Title = title
	})

	h.defaultCollector.OnHTML("html", func(html *colly.HTMLElement) {
		html.ForEach("h1,h2,h3,h4,h5,h6", func(index int, hElement *colly.HTMLElement) {
			page.HeadingsCount[hElement.Name]++
		})

		html.ForEach("input[type=password]", func(n int, hElement *colly.HTMLElement) {
			page.PasswordFieldCount[hElement.Name]++
		})

	})

	h.defaultCollector.OnResponse(func(r *colly.Response) {
		fmt.Println("visited", r.Ctx.Get("url"))
		// log.Println(string(r.Body))
		bytesReader := bytes.NewReader(r.Body)
		bufReader := bufio.NewReader(bytesReader)
		doctype, _, _ := bufReader.ReadLine()
		log.Println(string(doctype))
		page.HtmlVersion = string(doctype)
	})

	// Set error handler
	h.defaultCollector.OnError(func(r *colly.Response, err error) {
		fmt.Println("request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// record inaccessible links
	h.delegateCollector.OnError(func(r *colly.Response, err error) {
		fmt.Println("delegate: request URL:", r.Request.URL, "failed with response:", r.StatusCode, "\nError:", err.Error())

		// save all inaccessibleLinks
		links = append(links, model.Link{
			Url:          r.Request.URL.String(),
			StatusCode:   r.StatusCode,
			IsInternal:   false,
			IsAccessible: false,
		})
	})

	h.delegateCollector.OnResponse(func(r *colly.Response) {
		fmt.Println("delegate visited", r.Ctx.Get("url"))
		links = append(links, model.Link{
			Url:          r.Request.URL.String(),
			StatusCode:   r.StatusCode,
			IsInternal:   false,
			IsAccessible: true,
		})
	})

	h.defaultCollector.Visit(parsedURL.String())
	h.defaultCollector.Wait()

	page.Links = links
	c.JSON(http.StatusOK, gin.H{
		"results": page,
	})

	// h.storage.Client.Close()
	// delete previous data from storage
	if err := h.storage.Clear(); err != nil {
		log.Fatal(err)
	}
}
