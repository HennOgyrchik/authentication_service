package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	// чтение конфига
	conf, err := newConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//создание сервиса по конфигу
	service, err := newService(conf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//очистка бд от записей
	err = service.storage.dbCollection.collection.Drop(service.storage.dbCollection.ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	router := gin.Default()

	router.GET("/getTokens", service.getTokens)
	router.GET("/refreshToken", service.handRefresh)

	err = router.Run(service.addr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
