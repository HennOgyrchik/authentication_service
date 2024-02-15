package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	///////////////////////////////////////////// читать из файла
	address := "192.168.0.116:8080"
	key := "very-secret-key"
	aMinute := 1
	rMinute := 3
	bcryptCost := bcrypt.DefaultCost

	/////////////////////////////////////////////////

	service := newService(address, key, aMinute, rMinute, bcryptCost)

	//очистка бд от записей
	err := service.storage.dbCollection.collection.Drop(service.storage.dbCollection.ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	router := gin.Default()

	router.GET("/getTokens", service.getTokens)
	router.GET("/refresh", service.handRefresh)

	router.Run(service.addr)
}
