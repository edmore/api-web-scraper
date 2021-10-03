package resource

import (
	"net/http"
	"net/url"

	"github.com/edmore/api-web-scraper/model"
	"github.com/edmore/api-web-scraper/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	Collector service.CollectorService
}

func (h *Handler) GetPageContents(c *gin.Context) {
	u := c.Query("url")

	parsedURL, err := url.Parse(u)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if err := h.Collector.Reset(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	h.Collector.Init()
	h.Collector.Visit(parsedURL.String())

	c.JSON(http.StatusOK, gin.H{
		"results": model.Page{
			HtmlVersion:   h.Collector.GetHtmlVersion(),
			Title:         h.Collector.GetPageTitle(),
			HeadingsCount: h.Collector.GetHeadings(),
			Links:         h.Collector.GetLinks(),
		},
	})
}
