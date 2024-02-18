package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"medods/internal/api"
	"medods/internal/config"
)

func main() {
	// чтение конфига
	conf, err := config.NewConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//создание сервиса по конфигу
	service, err := api.NewService(conf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//очистка бд от записей
	err = service.ClearStorage()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	router := gin.Default()

	router.GET("/getTokens", service.GetTokens)
	router.GET("/refreshTokens", service.RefreshTokens)

	err = router.Run(conf.GetServiceAddress())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
