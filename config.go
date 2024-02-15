package main

import (
	"os"
	"strconv"
)

type config struct {
	serviceAddr         string
	secretKey           string
	expTimeAccessToken  int
	expTimeRefreshToken int
	bcryptCost          int
	dbAddr              string
}

func newConfig() (config, error) {
	serviceAddr, exists := os.LookupEnv("SERVICE_ADDRESS")
	if !exists {
		serviceAddr = ":8080"
	}

	secretKey, _ := os.LookupEnv("SECRET_KEY")
	if !exists {
		secretKey = "Default secret key"
	}

	aTime, exists := os.LookupEnv("EXPIRATION_TIME_ACCESS_TOKEN")
	var err error
	var expTimeAccessToken int
	if !exists {
		expTimeAccessToken = 10
	} else {
		expTimeAccessToken, err = strconv.Atoi(aTime)
		if err != nil {
			return config{}, err
		}
	}

	rTime, exists := os.LookupEnv("EXPIRATION_TIME_REFRESH_TOKEN")
	var expTimeRefreshToken int
	if !exists {
		expTimeRefreshToken = 30
	} else {
		expTimeRefreshToken, err = strconv.Atoi(rTime)
		if err != nil {
			return config{}, err
		}
	}

	cost, exists := os.LookupEnv("BCRYPT_COST")
	var bcryptCost int
	if !exists {
		bcryptCost = 10
	} else {
		bcryptCost, err = strconv.Atoi(cost)
		if err != nil {
			return config{}, err
		}
	}

	dbAddr, exists := os.LookupEnv("DATABASE_ADDRESS")
	if !exists {
		return config{}, ErrAdrrDBNotFound
	}

	return config{
		serviceAddr:         serviceAddr,
		secretKey:           secretKey,
		expTimeAccessToken:  expTimeAccessToken,
		expTimeRefreshToken: expTimeRefreshToken,
		bcryptCost:          bcryptCost,
		dbAddr:              dbAddr,
	}, nil

}
