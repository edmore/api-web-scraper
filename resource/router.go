package resource

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/gocolly/redisstorage"
)

func NewRouter(defaultCollector *colly.Collector, delegateCollector *colly.Collector, storage *redisstorage.Storage) *gin.Engine {
	router := gin.Default()
	basePath := "/scraper"

	handlers := Handler{defaultCollector, delegateCollector, storage}

	r := router.Group(basePath)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"results": "pong",
		})
	})

	r.GET("/page-contents", handlers.GetPageContents)

	return router
}
