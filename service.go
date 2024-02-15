package main

import (
	"encoding/base64"
	"errors"
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
	addr                       string
	jwtSecretKey               []byte
	expirationTimeAccessToken  time.Duration
	expirationTimeRefreshToken time.Duration
}

func newService() *service {
	return &service{
		///////////////////////////////////////////// читать из файла
		addr:                       "192.168.0.116:8080",
		jwtSecretKey:               []byte("very-secret-key"),
		expirationTimeAccessToken:  time.Duration(time.Hour * 12),
		expirationTimeRefreshToken: time.Duration(time.Hour * 12),
		storage:                    *newStorage(),
	}
}

func (s *service) getTokens(c *gin.Context) {
	//читаем guid
	var json struct {
		GUID string `json:"GUID" binding:"required"`
	}
	err := c.ShouldBindJSON(&json)
	if err != nil {
		s.sendError(c, http.StatusBadRequest, "The GUID Field Was Not Found")
		return
	}

	//проверка полученного из json GUID(строка) на соответствие типу данных guid.GUID, попыткой преобразования типа string -> guid.GUID
	tmp := []byte(json.GUID)
	tmpGUID := guid.GUID{}
	err = tmpGUID.UnmarshalText(tmp)
	if err != nil {
		s.sendError(c, http.StatusBadRequest, "Invalid Data")
		return
	}

	//Проверка прошла успешно, записываем GUID как строку.
	userGUID := json.GUID

	//ПРОВЕКА НА СУЩЕСТВОВАНИЕ!
	if _, ok := s.storage.accessMap[userGUID]; ok {
		s.sendError(c, http.StatusBadRequest, "Already Exists")
		return
	}
	//создаем новые токены
	accessToken, err := s.newAccessToken(userGUID)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	refreshToken := s.newRefreshToken()

	//запоминаем токены
	err = s.storage.rememberTokens(userInfo{
		guid:                userGUID,
		accessToken:         accessToken,
		expTimeAccessToken:  s.expirationTimeAccessToken,
		refreshToken:        refreshToken,
		expTimeRefreshToken: s.expirationTimeRefreshToken,
	})
	if err != nil {
		if errors.Is(err, errors.New(alreadyExists)) {
			s.sendError(c, http.StatusBadRequest, "Already Exists")
		} else {
			s.sendError(c, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	//пишем access токен в куки и отправляем токены json'ом
	c.SetCookie("user_cookie", accessToken, int(s.expirationTimeAccessToken), "/", "192.168.0.116", false, true)
	c.JSON(http.StatusOK, struct {
		Access_Token  string
		Refresh_Token string
	}{accessToken, refreshToken})

}

func (s *service) handRefresh(c *gin.Context) {
	//читаем рефреш токен из json
	var json struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	err := c.ShouldBindJSON(&json)
	if err != nil {
		c.JSON(http.StatusBadRequest, struct{}{})
		return
	}

	//проверяем есть ли такой refresh токен в БД
	rows, err := s.dbCollection.find(json.RefreshToken)
	if len(rows) != 0 {
		c.JSON(http.StatusBadRequest, struct{}{})
		return
	}
	//проверяем действителен ли еще токен

	//читаем куки
	accessToken, err := c.Cookie("user_cookie")
	if err != nil {
		c.JSON(http.StatusBadRequest, struct{}{})
	}
	fmt.Println("access token: ", accessToken)

	// проверяем связку refresh - access

	//выдаем новую пару

}

func (s *service) newAccessToken(guid string) (t string, err error) {
	payload := jwt.MapClaims{
		"guid": guid,
		"exp":  time.Now().Unix() + int64(s.expirationTimeAccessToken),
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

func (s *service) sendError(c *gin.Context, code int, message string) {
	c.JSON(code, struct {
		Message string
	}{message})
}
