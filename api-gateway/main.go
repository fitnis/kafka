package main

import (
	"github.com/fitnis/api-gateway/proxy"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	proxy.RegisterRoutes(router)
	router.Run(":8080") // Public API port
}
