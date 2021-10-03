package resource

import (
	"net/http"

	"github.com/edmore/api-web-scraper/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(srv service.CollectorService) *gin.Engine {
	router := gin.Default()
	basePath := "/scraper"

	handlers := Handler{srv}

	r := router.Group(basePath)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"results": "pong",
		})
	})

	r.GET("/page-contents", handlers.GetPageContents)

	return router
}
