package main

import "github.com/gin-gonic/gin"

func main() {
	service := newService()
	//это надо будет убрать!
	service.storage.dbCollection.collection.Drop(service.dbCollection.ctx)

	router := gin.Default()

	router.GET("/getTokens", service.getTokens)
	router.GET("/refresh", service.handRefresh)

	router.Run(service.addr)
}
