package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

type Controller struct {
	Rdb    *redis.Client
	DbPool *pgxpool.Pool
}

type RUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Resp struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

func (c Controller) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var ru RUser
	err := json.NewDecoder(r.Body).Decode(&ru)
	if err != nil || (ru.Password == "" || ru.Username == "") {
		if err == nil {
			err = errors.New("wrong request")
		}
		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
		return
	}
	newUserId := GetNewUUID()
	hashedPassword, _ := HashPassword(ru.Password)
	_, err = c.DbPool.Exec(context.Background(), "INSERT INTO users(userId,username,password) VALUES ($1,$2,$3)", newUserId, ru.Username, hashedPassword)
	if err != nil {
		sendResp(w, http.StatusInternalServerError, Resp{Message: err.Error()})
		return
	}
	sendResp(w, http.StatusCreated, Resp{
		Message: "user create successfully",
		Data:    map[string]interface{}{"userId": newUserId},
	})

}

func (c Controller) LoginUser(w http.ResponseWriter, r *http.Request) {
	var userId string
	var passwordHash string
	var ru RUser
	err := json.NewDecoder(r.Body).Decode(&ru)
	if err != nil || (ru.Password == "" || ru.Username == "") {
		if err == nil {
			err = errors.New("wrong request")
		}
		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
		return
	}
	err = c.DbPool.QueryRow(context.Background(), "SELECT (userId,password) from users where username=$1", ru.Username).Scan(&userId, &passwordHash)
	if err != nil {
		sendResp(w, http.StatusInternalServerError, Resp{Message: err.Error()})
		return
	}
	if ch := CheckPasswordHash(ru.Password, passwordHash); !ch {
		sendResp(w, http.StatusOK, Resp{Message: "incorrect username or password"})
		return
	}
	jwtToken, err := CreateJwtToken(ru.Username)
	if err != nil {
		sendResp(w, http.StatusInternalServerError, Resp{Message: err.Error()})
		return
	}
	sendResp(w, http.StatusOK, Resp{
		Message: "Logged in successfully",
		Data:    map[string]interface{}{"token": jwtToken, "userId": userId},
	})

}

func GetUserId(dbpool *pgxpool.Pool, username string) (string, error) {
	return username, nil
}
