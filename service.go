package main

import (
	"encoding/base64"
	"errors"
	"github.com/Microsoft/go-winio/pkg/guid"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type service struct {
	storage                    storage
	addr                       string
	jwtSecretKey               []byte
	expirationTimeAccessToken  int64
	expirationTimeRefreshToken int64
}

func newService(address string, key string, expTimeAccessTokenInMinute int, expTimeRefreshTokenInMinute int, bcryptCost int) *service {
	return &service{
		addr:                       address,
		jwtSecretKey:               []byte(key),
		expirationTimeAccessToken:  int64(time.Minute * time.Duration(expTimeAccessTokenInMinute)),
		expirationTimeRefreshToken: int64(time.Minute * time.Duration(expTimeRefreshTokenInMinute)),
		storage:                    *newStorage(bcryptCost),
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

	//проверка полученного из json GUID(строка)
	_, err = convertToGUID(json.GUID)
	if err != nil {
		s.sendError(c, http.StatusBadRequest, "Invalid Data")
	}

	//Создаем новую пару
	access, refresh, err := s.createAndRememberPairTokens(json.GUID)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
	}

	//////////////////////////////////пишем guid  в куки и отправляем токены json'ом
	c.SetCookie("user", json.GUID, int(s.expirationTimeAccessToken), "/", strings.Split(s.addr, ":")[0], false, true)

	//Отправляем пару
	c.JSON(http.StatusOK, struct {
		Access  string
		Refresh string
	}{access, refresh})

}

func (s *service) createAndRememberPairTokens(guid string) (accessToken, refreshToken string, err error) {

	//создаем новые токены
	accessToken, err = s.newAccessToken(guid)
	if err != nil {
		return
	}
	refreshToken = s.newRefreshToken()

	//запоминаем токены
	err = s.storage.rememberTokens(guid,
		tokenInfo{token: accessToken, expTime: time.Now().Unix() + s.expirationTimeAccessToken},
		tokenInfo{token: refreshToken, expTime: time.Now().Unix() + s.expirationTimeRefreshToken})
	return
}

func (s *service) handRefresh(c *gin.Context) {
	guid, err := c.Cookie("user")
	if err != nil {
		c.JSON(http.StatusBadRequest, struct{}{})
	}

	//читаем refresh токен из json
	var json struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	err = c.ShouldBindJSON(&json)
	if err != nil {
		s.sendError(c, http.StatusBadRequest, "The refresh_token Field Was Not Found")
		return
	}

	//проверяем есть ли такой refresh токен в БД
	hash, err := s.storage.findHash(guid, json.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			s.sendError(c, http.StatusBadRequest, ErrNotFound.Error())
		} else {
			s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
		}
		return
	}

	row, err := s.storage.findOne(hash)

	//проверяем действителен ли еще токен
	if row.ExpTime <= time.Now().Unix() {
		s.sendError(c, http.StatusBadRequest, ErrExpTimeHasExpired.Error())
		return
	}

	//удаляем старую пару
	err = s.storage.deleteToken(row.Guid, hash)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
	}

	//Создаем новую пару
	access, refresh, err := s.createAndRememberPairTokens(row.Guid)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
	}

	//Отправляем пару
	c.JSON(http.StatusOK, struct {
		Access  string
		Refresh string
	}{access, refresh})

}

// Сгенерировать новый Access Token
func (s *service) newAccessToken(guid string) (t string, err error) {
	payload := jwt.MapClaims{
		"guid": guid,
		"exp":  time.Now().Unix() + s.expirationTimeAccessToken,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, payload)

	t, err = token.SignedString(s.jwtSecretKey)

	return
}

// Сгенерировать новый Refresh Token
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

// Отправить JSON с сообщением об ошибке
func (s *service) sendError(c *gin.Context, code int, message string) {
	c.JSON(code, struct {
		Error string
	}{message})
}

// Проверка на соответствие типу данных guid.GUID попыткой преобразования типа string -> guid.GUID
func convertToGUID(str string) (guid.GUID, error) {
	tmp := []byte(str)
	tmpGUID := guid.GUID{}
	err := tmpGUID.UnmarshalText(tmp)
	return tmpGUID, err
}
