package main

import (
	"github.com/gin-gonic/gin"
)

// Specifies Stork backend version.
const VERSION = "0.0.1"

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/version-get", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": VERSION,
		})
	})

	return r
}

func main() {
	r := setupRouter()
	r.Run() // listen and serve on 0.0.0.0:8080
}
