package controllers

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

type MessageResp struct {
	SenderId   string    `json:"senderId"`
	ReceiverId string    `json:"receiverId"`
	Content    string    `json:"content"`
	SentAt     time.Time `json:"sentAt"`
	MessageId  string    `json:"messageId"`
}

func (c Controller) GetMessages(w http.ResponseWriter, r *http.Request) {
	ctxVal := ForContext(r.Context())
	if !ctxVal.Authenticated {
		sendResp(w, http.StatusUnauthorized, Resp{Message: "you are not authenticated"})
		return
	}
	vars := mux.Vars(r)
	friendId := vars["friendId"]
	messageNum := r.URL.Query().Get("num")
	num, err := strconv.Atoi(messageNum)
	if err != nil {
		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
		return
	}
	if friendId == "" || num <= 0 {
		sendResp(w, http.StatusBadRequest, Resp{Message: "incorrect url"})
		return
	}

	rows, err := c.DbPool.Query(context.Background(), "select senderId,receiverId,content,sentAt,messageId from messages offset $1 LIMIT $2 where (senderId=$3 or receiverId=$4) and (senderId=$5 or receiverId=$6) order by sentAt desc", num*20, 20, ctxVal.UserId, ctxVal.UserId, friendId, friendId)
	if err != nil {
		sendResp(w, http.StatusInternalServerError, Resp{Message: err.Error()})
		return
	}
	var msr []MessageResp
	for rows.Next() {
		var senderId, receiverId, content, messageId string
		var sentAt time.Time
		err = rows.Scan(&senderId, &receiverId, &content, &sentAt, &messageId)
		if err != nil {
			sendResp(w, http.StatusInternalServerError, Resp{Message: err.Error()})
			return
		}
		msr = append(msr, MessageResp{
			SenderId:   senderId,
			ReceiverId: receiverId,
			Content:    content,
			SentAt:     sentAt,
			MessageId:  messageId,
		})

	}
	sendResp(w, http.StatusOK, Resp{
		Message: "all messages fetched",
		Data:    map[string]interface{}{"messages": msr},
	})
}
