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
	storage                    *storage
	addr                       string
	jwtSecretKey               []byte
	expirationTimeAccessToken  int64
	expirationTimeRefreshToken int64
}

// newService Создание нового сервиса, заполнение полей
func newService(config config) (*service, error) {
	addr := strings.Split(config.dbAddr, ":")

	store, err := newStorage(config.bcryptCost, addr[0], addr[1])
	if err != nil {
		return nil, err
	}

	return &service{
		addr:                       config.serviceAddr,
		jwtSecretKey:               []byte(config.secretKey),
		expirationTimeAccessToken:  int64(time.Minute * time.Duration(config.expTimeAccessToken)),
		expirationTimeRefreshToken: int64(time.Minute * time.Duration(config.expTimeRefreshToken)),
		storage:                    store,
	}, err
}

// getTokens обработчик "/getTokens"
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
	_, err = stringToGUID(json.GUID)
	if err != nil {
		s.sendError(c, http.StatusBadRequest, "Invalid Data")
	}

	//Создаем новую пару
	access, refresh, err := s.createAndRememberPairTokens(json.GUID)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
	}

	//пишем guid  в куки
	c.SetCookie("user", json.GUID, int(s.expirationTimeRefreshToken), "/", strings.Split(s.addr, ":")[0], false, true)

	//Отправляем пару json'ом
	c.JSON(http.StatusOK, struct {
		Access  string
		Refresh string
	}{access, refresh})

}

// createAndRememberPairTokens создание пары Access-Refresh, запись в БД и accessMap
func (s *service) createAndRememberPairTokens(guid string) (accessToken, refreshToken string, err error) {

	// создаем новые токены
	accessToken, err = s.newAccessToken(guid)
	if err != nil {
		return
	}
	refreshToken = s.newRefreshToken()

	// запоминаем токены
	err = s.storage.rememberTokens(guid,
		tokenInfo{token: accessToken, expTime: time.Now().Unix() + s.expirationTimeAccessToken},
		tokenInfo{token: refreshToken, expTime: time.Now().Unix() + s.expirationTimeRefreshToken})
	return
}

// handRefresh обработчик "/refreshToken"
func (s *service) handRefresh(c *gin.Context) {
	uGuid, err := c.Cookie("user")
	if err != nil {
		s.sendError(c, http.StatusBadRequest, "bad request")
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
	hash, err := s.storage.findHash(uGuid, json.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			s.sendError(c, http.StatusBadRequest, ErrNotFound.Error())
		} else {
			s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
		}
		return
	}

	row, err := s.storage.dbCollection.findOne(hash)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
	}

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

// newAccessToken генерация нового Access Token
func (s *service) newAccessToken(guid string) (t string, err error) {
	payload := jwt.MapClaims{
		"guid": guid,
		"exp":  time.Now().Unix() + s.expirationTimeAccessToken,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, payload)

	return token.SignedString(s.jwtSecretKey)
}

// newRefreshToken генерация нового Refresh Token
func (s *service) newRefreshToken() string {
	var charset = []byte(base64Alphabet)
	salt := time.Now().Format(timeLayout)
	n := rand.Intn(50)
	randStr := make([]byte, n, n+len(salt))

	for i := range randStr {
		randStr[i] = charset[rand.Intn(len(charset))]
	}

	randStr = append(randStr, salt[:]...)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(randStr), func(i, j int) { randStr[i], randStr[j] = randStr[j], randStr[i] })

	return base64.StdEncoding.EncodeToString(randStr)
}

// sendError отправка JSON с сообщением об ошибке
func (s *service) sendError(c *gin.Context, code int, message string) {
	c.JSON(code, struct {
		Error string
	}{message})
}

// convertToGUID Проверка на соответствие типу данных guid.GUID попыткой преобразования string -> guid.GUID
func stringToGUID(str string) (guid.GUID, error) {
	tmp := []byte(str)
	tmpGUID := guid.GUID{}
	err := tmpGUID.UnmarshalText(tmp)
	return tmpGUID, err
}
