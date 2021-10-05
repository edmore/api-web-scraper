package resource

import (
	"net/http"
	"net/url"

	"github.com/edmore/api-web-scraper/model"
	"github.com/edmore/api-web-scraper/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	CollectorService service.CollectorService
}

func (h *Handler) GetPageContents(c *gin.Context) {
	u := c.Query("url")
	parsedURL, err := url.Parse(u)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if err := h.CollectorService.Reset(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	h.CollectorService.Visit(parsedURL.String())

	links, linksCount := h.CollectorService.GetLinks()
	c.JSON(http.StatusOK, gin.H{
		"results": model.Page{
			HtmlVersion:   h.CollectorService.GetHtmlVersion(),
			Title:         h.CollectorService.GetPageTitle(),
			HeadingsCount: h.CollectorService.GetHeadings(),
			Links:         links,
			LinksCount:    linksCount,
			HasLoginForm:  h.CollectorService.HasLoginForm(),
		},
	})
}
