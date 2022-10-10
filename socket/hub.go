package socket

import (
	"context"
	"fmt"
	"github.com/ThunderGod77/nessenger/controllers"
	"github.com/go-redis/redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"strconv"
	"time"
)

type ClientInfo struct {
	UserId string
	Ct     *Client
}

type PM struct {
	Message    string `json:"message"`
	ReceiverId string `json:"receiverId"`
	SenderId   string `json:"senderId"`
}

type Hub struct {
	Users            map[string]*Client
	Clients          map[*Client]string
	RegisterUser     chan *ClientInfo
	RegisterClient   chan *Client
	UnregisterClient chan *Client
	UnregisterUser   chan string
	Rdb              *redis.Client
	DbPool           *pgxpool.Pool
}

func NewHub(rdb *redis.Client, dbPool *pgxpool.Pool) *Hub {
	return &Hub{
		RegisterClient:   make(chan *Client),
		RegisterUser:     make(chan *ClientInfo),
		UnregisterUser:   make(chan string),
		UnregisterClient: make(chan *Client),
		Clients:          make(map[*Client]string),
		Rdb:              rdb,
		DbPool:           dbPool,
	}
}
func AddOnlineStatus(rdb *redis.Client, userId string, cTime time.Time) error {
	hour, min, _ := cTime.Clock()
	ts := strconv.Itoa(hour) + strconv.Itoa((min/5)*5)
	cmd := rdb.SAdd(context.Background(), fmt.Sprintf("players:Online:%s", ts), userId)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}
func (h *Hub) AddOnlineStatus(userId string) {
	err := AddOnlineStatus(h.Rdb, userId, time.Now())
	if err != nil {
		log.Println(err)
		return
	}
	err = AddOnlineStatus(h.Rdb, userId, time.Now().Add(time.Minute*5))
	if err != nil {
		log.Println(err)
		return
	}
}
func (h *Hub) StoreFriends(userId string) {
	rows, err := h.DbPool.Query(context.Background(), "SELECT users.userId,users.username,fr.username,fr.userId FROM friends JOIN users ON friends.userid=users.userId JOIN users AS fr ON fr.userId=friends.friendId WHERE (friends.userId=$1 OR friends.friendId=$2) AND status=$3", userId, userId, "accepted")
	if err != nil {
		log.Println(err)
		return
	}
	var ans []interface{}
	for rows.Next() {
		var usrId, username string
		var friendId, friendName string
		err = rows.Scan(&usrId, &username, &friendId, &friendName)
		if err != nil {
			log.Println(err)
			return
		}
		if usrId == usrId {
			ans = append(ans, friendId)
		} else {
			ans = append(ans, usrId)
		}

	}
	err = h.Rdb.SAdd(context.Background(), userId+":friends", ans...).Err()
	if err != nil {
		log.Println(err)
		return
	}
}
func AddOfflineStatus() {}
func AddMessageToDataBase(dbPool *pgxpool.Pool, pm PM) {
	newMessageId := controllers.GetNewUUID()

	_, err := dbPool.Exec(context.Background(), "insert into messages(messageId,senderId,receiverId,content,sentAt) values ($1,$2,$3,$4,$5)", newMessageId, pm.SenderId, pm.ReceiverId, pm.Message, time.Now())
	if err != nil {
		log.Println(err)
		return
	}

}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.RegisterClient:
			h.Clients[c] = "connected"
		case u := <-h.RegisterUser:
			h.Clients[u.Ct] = u.UserId
			h.Users[u.UserId] = u.Ct
			h.AddOnlineStatus(u.UserId)
			h.StoreFriends(u.UserId)
		case c := <-h.UnregisterUser:
			client, ok := h.Users[c]
			if !ok {
				log.Println("User did not exist already")
			} else {
				delete(h.Clients, client)
				delete(h.Users, c)
				AddOfflineStatus()
			}
		case cl := <-h.UnregisterClient:
			userId, ok := h.Clients[cl]
			if !ok || userId == "connected" {
				delete(h.Clients, cl)
				return
			}
			delete(h.Users, userId)
			delete(h.Clients, cl)
			AddOfflineStatus()

		}

	}
}

func (h *Hub) message(pm PM) {
	receiverClient, ok := h.Users[pm.ReceiverId]
	if !ok {
		AddMessageToDataBase(h.DbPool, pm)
		return
	}

	go func() {
		receiverClient.Send <- pm
	}()
	AddMessageToDataBase(h.DbPool, pm)

}
