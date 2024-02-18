package api

import (
	"encoding/base64"
	"errors"
	"github.com/Microsoft/go-winio/pkg/guid"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"math/rand"
	"medods/internal/config"
	"medods/internal/storage"
	"net/http"
	"strings"
	"time"
)

type Service struct {
	storage                    *storage.Storage
	addr                       string
	jwtSecretKey               []byte
	expirationTimeAccessToken  int64
	expirationTimeRefreshToken int64
}

const base64Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/"
const timeLayout = "200612345"

// NewService Создание нового сервиса, заполнение полей
func NewService(config *config.Config) (*Service, error) {
	addr := strings.Split(config.GetDBAddress(), ":")

	store, err := storage.NewStorage(config.GetBcryptCost(), addr[0], addr[1])
	if err != nil {
		return nil, err
	}

	return &Service{
		addr:                       config.GetServiceAddress(),
		jwtSecretKey:               []byte(config.GetSecretKey()),
		expirationTimeAccessToken:  int64(time.Minute * time.Duration(config.GetExpTimeAccessToken())),
		expirationTimeRefreshToken: int64(time.Minute * time.Duration(config.GetExpTimeRefreshToken())),
		storage:                    store,
	}, err
}

// GetTokens обработчик "/getTokens"
func (s *Service) GetTokens(c *gin.Context) {
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
		return
	}

	//Создаем новую пару
	access, refresh, err := s.createAndRememberPairTokens(json.GUID)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
		return
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
func (s *Service) createAndRememberPairTokens(guid string) (accessToken, refreshToken string, err error) {

	// создаем новые токены
	accessToken, err = s.newAccessToken(guid)
	if err != nil {
		return
	}
	refreshToken = s.newRefreshToken()

	// запоминаем токены
	err = s.storage.RememberTokens(
		storage.TokenInfo{Token: accessToken, ExpTime: time.Now().Unix() + s.expirationTimeAccessToken, GUID: guid},
		storage.TokenInfo{Token: refreshToken, ExpTime: time.Now().Unix() + s.expirationTimeRefreshToken, GUID: guid})
	return
}

// RefreshTokens обработчик "/refreshTokens"
func (s *Service) RefreshTokens(c *gin.Context) {
	uGuid, err := c.Cookie("user")
	if err != nil {
		s.sendError(c, http.StatusBadRequest, "bad request")
		return
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
	hash, err := s.storage.FindHash(uGuid, json.RefreshToken)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			s.sendError(c, http.StatusBadRequest, storage.ErrNotFound.Error())
		} else {
			s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
		}
		return
	}

	row, err := s.storage.FindOneInDB(hash)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
		return
	}

	//проверяем действителен ли еще токен
	if row.ExpTime <= time.Now().Unix() {
		s.sendError(c, http.StatusBadRequest, ErrExpTimeHasExpired.Error())
		return
	}

	//удаляем старую пару
	err = s.storage.DeleteToken(row.GUID, hash)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
		return
	}

	//Создаем новую пару
	access, refresh, err := s.createAndRememberPairTokens(row.GUID)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, InternalServerError.Error())
		return
	}

	//Отправляем пару
	c.JSON(http.StatusOK, struct {
		Access  string
		Refresh string
	}{access, refresh})

}

// newAccessToken генерация нового Access Token
func (s *Service) newAccessToken(guid string) (t string, err error) {
	payload := jwt.MapClaims{
		"guid": guid,
		"exp":  time.Now().Unix() + s.expirationTimeAccessToken,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, payload)

	return token.SignedString(s.jwtSecretKey)
}

// newRefreshToken генерация нового Refresh Token
func (s *Service) newRefreshToken() string {
	var charset = []byte(base64Alphabet)
	salt := time.Now().Format(timeLayout)

	n := rand.Intn(40)
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
func (s *Service) sendError(c *gin.Context, code int, message string) {
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

// ClearStorage очистка хранилищ от данных
func (s *Service) ClearStorage() error {
	return s.storage.ClearStorage()
}
