package main

import (
	"encoding/base64"
	"fmt"
	"github.com/Microsoft/go-winio/pkg/guid"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"math/rand"
	"net/http"
	"time"
)

type service struct {
	storage
	addr           string
	jwtSecretKey   []byte
	expirationTime int
}

type Id struct {
	GUID string `json:"GUID" binding:"required"`
}

func newService() *service {
	return &service{
		///////////////////////////////////////////// читать из файла
		addr:           "192.168.0.116:8080",
		jwtSecretKey:   []byte("very-secret-key"),
		expirationTime: int(time.Hour * 12),
		storage: storage{
			accessMap: make(map[guid.GUID]string),
		},
	}
}

func (s *service) getTokens(c *gin.Context) {
	//читаем guid
	var id Id
	err := c.ShouldBindJSON(&id)
	if err != nil {
		fmt.Println(err, id.GUID)
		c.JSON(http.StatusBadRequest, struct{}{})
		return
	}
	b := []byte(id.GUID)

	userGUID := guid.GUID{}
	err = userGUID.UnmarshalText(b)
	if err != nil {
		fmt.Println("не гуид!")
		c.JSON(http.StatusBadRequest, struct{}{})
		return
	}

	//проверяем существуют ли токены для этого guid. если уже есть, вернуть ошибку 400 и выйти.
	if _, ok := s.accessMap[userGUID]; ok {
		c.JSON(http.StatusBadRequest, struct{}{})
		return
	}

	//создаем новые токены
	accessToken, err := s.newAccessToken(userGUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, struct{}{})
		return
	}
	refreshToken := s.newRefreshToken()

	//запоминаем токены
	err = s.storage.rememberTokens(userGUID, accessToken, refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, struct{}{})
		return
	}

	//пишем access токен в куки и отправляем токены json'ом
	c.SetCookie("user_cookie", accessToken, s.expirationTime, "/", "192.168.0.116", false, true)
	c.JSON(http.StatusOK, struct {
		Access_Token  string
		Refresh_Token string
	}{accessToken, refreshToken})

}

func (s *service) handRefresh(c *gin.Context) {
	//var refreshToken string

	cookie, err := c.Cookie("user_cookie")
	if err != nil {
		c.JSON(http.StatusBadRequest, struct{}{})
	}
	fmt.Println(cookie)

	// надо искать в монге

}

func (s *service) newAccessToken(guid guid.GUID) (t string, err error) {
	payload := jwt.MapClaims{
		"guid": guid,
		"exp":  s.expirationTime,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, payload)

	t, err = token.SignedString(s.jwtSecretKey)

	return
}

func (s *service) newRefreshToken() string {
	var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/")
	salt := time.Now().Format("200612345")
	n := rand.Intn(50)
	randStr := make([]byte, n, n+len(salt))

	for i := range randStr {
		randStr[i] = charset[rand.Intn(len(charset))]
	}

	randStr = append(randStr, salt[:]...)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(randStr), func(i, j int) { randStr[i], randStr[j] = randStr[j], randStr[i] })

	return base64.StdEncoding.EncodeToString(randStr)
}
