package main

import (
	"context"
	"github.com/ThunderGod77/nessenger/controllers"
	"github.com/ThunderGod77/nessenger/socket"
	"github.com/go-redis/redis/v9"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"

	"log"
	"net/http"
)

func authMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := r.Header.Get("authorization")
			if c == "" {
				ctx := context.WithValue(r.Context(), controllers.UserCtxKey, controllers.User{
					Authenticated: false,
				})

				// and call the next with our new context
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)

				return
			}

			// Allow unauthenticated users in
			username, err := controllers.VerifyJwtToken(c)
			if err != nil {
				http.Error(w, "Invalid authorization token", http.StatusForbidden)
				return
			}

			// get the user from the database
			userId, err := controllers.GetUserId(pool, username)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			// put it in context
			ctx := context.WithValue(r.Context(), controllers.UserCtxKey, controllers.User{
				Username:      username,
				UserId:        userId,
				Authenticated: true,
			})

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func main() {

	dbPool, err := pgxpool.New(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81", // no password set
		DB:       1,                                  // use default DB
	})
	h := socket.NewHub(rdb, dbPool)
	go h.Run()

	cntrl := controllers.Controller{
		Rdb:    rdb,
		DbPool: dbPool,
	}

	r := mux.NewRouter()
	u := r.PathPrefix("/user").Subrouter()
	u.HandleFunc("/", cntrl.RegisterUser).Methods("POST")
	u.HandleFunc("/login", cntrl.LoginUser).Methods("POST")

	f := r.PathPrefix("/friends").Subrouter()
	f.Use(authMiddleware(dbPool))
	f.HandleFunc("/", cntrl.GetFriends).Methods("GET")
	f.HandleFunc("/request", cntrl.SendFriendRequest).Methods("POST")
	f.HandleFunc("/request", cntrl.AcceptFriendRequest).Methods("PUT")
	f.HandleFunc("/request", cntrl.DeclineFriendRequest).Methods("DELETE")
	f.HandleFunc("/request", cntrl.GetFriendRequests).Methods("GET")
	f.HandleFunc("/online", cntrl.GetFriendRequests)
	r.HandleFunc("/message/{friendId}", cntrl.GetMessages).Methods("GET")

	//r.HandleFunc("/ws", socket.WsHandler)
	log.Println("connect to http://localhost:8080/ ")
	log.Fatal(http.ListenAndServe(":8080", r))
}
