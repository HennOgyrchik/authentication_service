package main

import (
	"encoding/base64"
	"fmt"
	"github.com/Microsoft/go-winio/pkg/guid"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-jwt/jwt/v5"
	"math/rand"
	"net/http"
	"time"
)

// читать из файла
var jwtSecretKey = []byte("very-secret-key")

func main() {
	//вынести в аргументы
	var addr = "192.168.0.116:8080"
	////////

	router := gin.Default()

	router.GET("/getToken", handGet)
	router.GET("/refreshToken", handRefresh)
	//router.GET("/", newRefreshToken)
	router.Run(addr)
}

type request struct {
	GUID guid.GUID
}

func handGet(c *gin.Context) {
	//читаем json. обработать ошибку если прислали не GUID или что-то другое
	var json request
	err := c.BindJSON(&json)
	if err != nil {
		fmt.Println(err) //убрать
		return
	}

	accessToken, err := newAccessToken(json.GUID)
	if err != nil { //обработать ошибку "JWT token signing"
		return
	}
	refreshToken := newRefreshToken()

	c.JSON(http.StatusOK, struct{ Text string }{fmt.Sprintf("Hello %s", json.GUID)})
}

func handRefresh(c *gin.Context) {

}

func newAccessToken(guid guid.GUID) (t string, err error) {
	payload := jwt.MapClaims{
		"guid": guid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, payload)

	t, err = token.SignedString(jwtSecretKey)

	return
}

func newRefreshToken() string {
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
