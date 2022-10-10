package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

var mySigningKey = []byte("supersupersecretjwt singihke kljdfsoaihf iopsdfhyiouadsfuiogiasjigsdkjo;lfias")

var UserCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

type User struct {
	Username      string
	UserId        string
	Authenticated bool
}

type MyCustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetNewUUID() string {
	uid, _ := uuid.NewV4()

	return uid.String()
}

func sendResp(w http.ResponseWriter, statusCode int, resp interface{}) {
	log.Println(resp)
	respjs, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(respjs)
}

func CreateJwtToken(username string) (string, error) {
	claims := MyCustomClaims{
		username,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "test",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	return ss, err
}
func VerifyJwtToken(token string) (string, error) {
	vToken, err := jwt.ParseWithClaims(token, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := vToken.Claims.(*MyCustomClaims); ok && vToken.Valid {
		return claims.Username, nil
	}
	return "", errors.New("wrong jwt")
}
func ForContext(ctx context.Context) *User {
	raw, _ := ctx.Value(UserCtxKey).(*User)
	return raw
}
