package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type SFR struct {
	FriendId string `json:"friendId"`
}

type GCl struct {
	SenderId     string    `json:"senderId"`
	ReceiverId   string    `json:"receiverId"`
	SenderName   string    `json:"senderName"`
	ReceiverName string    `json:"receiverName"`
	SentAt       time.Time `json:"sentAt"`
	Content      string    `json:"content"`
}

func (c Controller) GetChatList(w http.ResponseWriter, r *http.Request) {
	//ctxVal := ForContext(r.Context())
	//if !ctxVal.Authenticated {
	//	sendResp(w, http.StatusUnauthorized, Resp{Message: "you are not authenticated"})
	//	return
	//}
	//rows, err := c.DbPool.Query(context.Background(), "select greatest(messages.senderId,messages.receiverId),least(messages.senderId,messages.receiverId),sr.username,rc.username,MAX(sentAt),content from messages join users as sr ON sr.userId=messages.senderId JOIN users as rc on rc.userId=messages.receiverId where senderId=$1 or receiverId=$2 group by greatest(messages.senderId,messages.receiverId),least(messages.senderId,messages.receiverId)", ctxVal.UserId, ctxVal.UserId)
	//if err != nil {
	//	sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
	//	return
	//}
	//
	//for rows.Next() {
	//	var usrId, username string
	//	var friendId, friendName string
	//	var content string
	//	var sentAt time.Time
	//	err = rows.Scan(&usrId, &username, &friendId, &friendName)
	//	if err != nil {
	//		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
	//		return
	//	}
	//	if usrId == ctxVal.UserId {
	//		ans[friendId] = friendName
	//	} else {
	//		ans[usrId] = username
	//	}
	//
	//}

}

func (c Controller) GetFriends(w http.ResponseWriter, r *http.Request) {
	ctxVal := ForContext(r.Context())
	if !ctxVal.Authenticated {
		sendResp(w, http.StatusUnauthorized, Resp{Message: "you are not authenticated"})
		return
	}
	rows, err := c.DbPool.Query(context.Background(), "SELECT users.userId,users.username,fr.username,fr.userId FROM friends JOIN users ON friends.userid=users.userId JOIN users AS fr ON fr.userId=friends.friendId WHERE (friends.userId=$1 OR friends.friendId=$2) AND status=$3", ctxVal.UserId, ctxVal.UserId, "accepted")
	if err != nil {
		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
		return
	}
	var ans map[string]string
	for rows.Next() {
		var usrId, username string
		var friendId, friendName string
		err = rows.Scan(&usrId, &username, &friendId, &friendName)
		if err != nil {
			sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
			return
		}
		if usrId == ctxVal.UserId {
			ans[friendId] = friendName
		} else {
			ans[usrId] = username
		}

	}
	sendResp(w, http.StatusCreated, Resp{
		Message: "Friends list",
		Data:    map[string]interface{}{"friends": ans},
	})

}

func (c Controller) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctxVal := ForContext(r.Context())
	if !ctxVal.Authenticated {
		sendResp(w, http.StatusUnauthorized, Resp{Message: "you are not authenticated"})
		return
	}
	var sfr SFR
	err := json.NewDecoder(r.Body).Decode(&sfr)
	if err != nil {
		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
		return
	}
	_, err = c.DbPool.Exec(context.Background(), "INSERT INTO FRIENDS (userId,friendId,status) values ($1,$2,$3)", ctxVal.UserId, sfr.FriendId, "request")
	if err != nil {
		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
		return
	}
	sendResp(w, http.StatusCreated, Resp{
		Message: "Sent friend request successfully",
	})
}

func (c Controller) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctxVal := ForContext(r.Context())
	if !ctxVal.Authenticated {
		sendResp(w, http.StatusUnauthorized, Resp{Message: "you are not authenticated"})
		return
	}
	var sfr SFR
	err := json.NewDecoder(r.Body).Decode(&sfr)
	if err != nil {
		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
		return
	}
	_, err = c.DbPool.Exec(context.Background(), "UPDATE FRIENDS SET status = $1 where userId=$2 and friendId=$3", sfr.FriendId, ctxVal.UserId, "accepted")
	if err != nil {
		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
		return
	}
	sendResp(w, http.StatusCreated, Resp{
		Message: "Accepted friend request successfully",
	})
}

func (c Controller) DeclineFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctxVal := ForContext(r.Context())
	if !ctxVal.Authenticated {
		sendResp(w, http.StatusUnauthorized, Resp{Message: "you are not authenticated"})
		return
	}
	var sfr SFR
	err := json.NewDecoder(r.Body).Decode(&sfr)
	if err != nil {
		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
		return
	}
	_, err = c.DbPool.Exec(context.Background(), "DELETE FROM FRIENDS where userId=$1 and friendId=$2", sfr.FriendId, ctxVal.UserId)
	if err != nil {
		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
		return
	}
	sendResp(w, http.StatusCreated, Resp{
		Message: "Declined friend request successfully",
	})
}

func (c Controller) GetFriendRequests(w http.ResponseWriter, r *http.Request) {
	ctxVal := ForContext(r.Context())
	if !ctxVal.Authenticated {
		sendResp(w, http.StatusUnauthorized, Resp{Message: "you are not authenticated"})
		return
	}
	rows, err := c.DbPool.Query(context.Background(), "SELECT users.userId,users.username FROM friends JOIN users ON friends.userId=users.userId WHERE friends.friendId=$1 AND status=$2", ctxVal.UserId, "request")
	if err != nil {
		sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
		return
	}
	var ans map[string]string
	for rows.Next() {
		var fId string
		var fName string
		err = rows.Scan(&fId, &fName)
		if err != nil {
			sendResp(w, http.StatusBadRequest, Resp{Message: err.Error()})
			return
		}
		ans[fId] = fName
	}
	sendResp(w, http.StatusCreated, Resp{
		Message: "friend requests list",
		Data:    map[string]interface{}{"requests": ans},
	})
}

func (c Controller) GetOnlineFriends(w http.ResponseWriter, r *http.Request) {
	ctxVal := ForContext(r.Context())
	if !ctxVal.Authenticated {
		sendResp(w, http.StatusUnauthorized, Resp{Message: "you are not authenticated"})
		return
	}
	hour, min, _ := time.Now().Clock()
	ts := strconv.Itoa(hour) + strconv.Itoa((min/5)*5)
	cmt := c.Rdb.SInter(context.Background(), fmt.Sprintf("players:Online:%s", ts), ctxVal.UserId+":friends")
	if cmt.Err() != nil {
		sendResp(w, http.StatusInternalServerError, Resp{Message: cmt.Err().Error()})
		return
	}
	val := cmt.Val()
	sendResp(w, http.StatusCreated, Resp{
		Message: "friend requests list",
		Data:    map[string]interface{}{"onlineFriends": val},
	})

}
