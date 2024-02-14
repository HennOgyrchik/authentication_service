package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	service := newService()

	router := gin.Default()

	router.GET("/getTokens", service.getTokens) //а точно ли так называются маршруты?
	router.GET("/refreshToken", service.handRefresh)

	router.Run(service.addr)
}
