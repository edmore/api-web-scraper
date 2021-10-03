package resource

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/edmore/api-web-scraper/model"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/gocolly/redisstorage"
)

type Handler struct {
	DefaultCollector  *colly.Collector
	DelegateCollector *colly.Collector
	Storage           *redisstorage.Storage
}

var doctypes = make(map[string]string)

func (h *Handler) GetPageContents(c *gin.Context) {
	u := c.Query("url")
	// validate URI
	parsedURL, err := url.Parse(u)
	if err != nil {
		// return error payload
		log.Fatal("oops")
	}
	// h.storage.Client.Set("url", parsedURL, 0)

	page := model.Page{HeadingsCount: make(map[string]int), PasswordFieldCount: make(map[string]int)}
	links := []model.Link{}

	doctypes["HTML 4.01 Strict"] = `"-//W3C//DTD HTML 4.01//EN"`
	doctypes["HTML 4.01 Transitional"] = `"-//W3C//DTD HTML 4.01 Transitional//EN"`
	doctypes["HTML 4.01 Frameset"] = `"-//W3C//DTD HTML 4.01 Frameset//EN"`
	doctypes["XHTML 1.0 Strict"] = `"-//W3C//DTD XHTML 1.0 Strict//EN"`
	doctypes["XHTML 1.0 Transitional"] = `"-//W3C//DTD XHTML 1.0 Transitional//EN"`
	doctypes["XHTML 1.0 Frameset"] = `"-//W3C//DTD XHTML 1.0 Frameset//EN"`
	doctypes["XHTML 1.1"] = `"-//W3C//DTD XHTML 1.1//EN"`
	doctypes["HTML 5"] = `<!DOCTYPE html>`

	// h.DefaultCollector.OnRequest(func(r *colly.Request) {
	// 	r.Ctx.Put("url", r.URL.String())
	// 	log.Println("visiting ...", r.URL.String())
	// })

	// On every a element which has href attribute call callback
	h.DefaultCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		fmt.Printf("link found: %q -> %s\n", e.Text, link)
		if link != "" {
			h.DelegateCollector.Visit(link)
			h.DelegateCollector.Wait()
		}
	})

	h.DefaultCollector.OnHTML("title", func(e *colly.HTMLElement) {
		title := e.Text
		page.Title = title
	})

	h.DefaultCollector.OnHTML("html", func(html *colly.HTMLElement) {
		html.ForEach("h1,h2,h3,h4,h5,h6", func(index int, hElement *colly.HTMLElement) {
			page.HeadingsCount[hElement.Name]++
		})

		html.ForEach("input[type=password]", func(n int, hElement *colly.HTMLElement) {
			page.PasswordFieldCount[hElement.Name]++
		})

	})

	h.DefaultCollector.OnResponse(func(r *colly.Response) {
		fmt.Println("visited", r.Ctx.Get("url"))
		// log.Println(string(r.Body))
		bytesReader := bytes.NewReader(r.Body)
		bufReader := bufio.NewReader(bytesReader)
		doctype, _, _ := bufReader.ReadLine()
		log.Println(string(doctype))
		page.HtmlVersion = determineDoctype(string(doctype))
	})

	// Set error handler
	// h.DefaultCollector.OnError(func(r *colly.Response, err error) {
	// 	fmt.Println("request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	// })

	// record inaccessible links
	h.DelegateCollector.OnError(func(r *colly.Response, err error) {
		fmt.Println("delegate: request URL:", r.Request.URL, "failed with response:", r.StatusCode, "\nError:", err.Error())

		// save all inaccessibleLinks
		links = append(links, model.Link{
			Url:          r.Request.URL.String(),
			StatusCode:   r.StatusCode,
			IsInternal:   false,
			IsAccessible: false,
		})
	})

	h.DelegateCollector.OnResponse(func(r *colly.Response) {
		fmt.Println("delegate visited", r.Ctx.Get("url"))
		links = append(links, model.Link{
			Url:          r.Request.URL.String(),
			StatusCode:   r.StatusCode,
			IsInternal:   false,
			IsAccessible: true,
		})
	})

	h.DefaultCollector.Visit(parsedURL.String())
	h.DefaultCollector.Wait()

	page.Links = links
	c.JSON(http.StatusOK, gin.H{
		"results": page,
	})

	// h.storage.Client.Close()
	// delete previous data from storage
	if err := h.Storage.Clear(); err != nil {
		log.Fatal(err)
	}
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
