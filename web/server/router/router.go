package router

import (
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/web/server/api/config"
)

func Router(r *gin.Engine) *gin.Engine {
	r.POST("/config/client", config.ClientConfig)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	return r
}
