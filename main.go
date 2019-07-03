package main

import (
  "database/sql"
  "log"
  "net/http"
  "os"
  "time"

  "github.com/joho/godotenv"
  "github.com/julienschmidt/httprouter"
  "github.com/lib/pq"
  "github.com/go-redis/redis"
)

var listen string
var postgres string
var redisHost string

var redisClient *redis.Client

func main() {
  // Load .env
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }
  listen = os.Getenv("LISTEN")
  postgres = os.Getenv("POSTGRES")
  redisHost = os.Getenv("REDIS")

  // Redis
  redisClient = redis.NewClient(&redis.Options{
    Addr: redisHost,
    Password: "",
    DB: 2,
  })

  // Postgres
  log.Printf("connecting to postgres %s", postgres)
  db, err := sql.Open("postgres", postgres)
  if err != nil {
    log.Fatal(err)
  }

  // Populate cache
  rows, err := db.Query(`
    SELECT user, conversation FROM "member"
  `)
  if err != nil {
    log.Fatal("Error retrieving records from database")
  }
  for rows.Next() {
    var userID, conversationID string
    if err := rows.Scan(&userID, &conversationID); err != nil {
      log.Fatal("Error retrieving records from database")
    }
    id := userID + "+" + conversationID
    redisClient.Set(id, true, 0)
  }
  rows.Close()
  db.Close()

  // Start cache update listener
  minReconn := 10 * time.Second
  maxReconn := 1 * time.Minute
  listener := pq.NewListener(postgres, minReconn, maxReconn, func(ev pq.ListenerEventType, err error) {
    if err != nil {
      log.Fatal(err)
    } else if ev == pq.ListenerEventConnected {
      log.Println("listener connected")
    }
  })

  // INSERT/UPDATE Listener
  err = listener.Listen("member_new")
  if err != nil {
    log.Fatal(err)
  }

  // DELETE Listener
  err = listener.Listen("member_delete")
  if err != nil {
    log.Fatal(err)
  }

  // Process events
  go ListenForEvents(listener)

  // Routes
  router := httprouter.New()
  router.GET("/user/:userid/conversation/:conversationid", GetPermission)

  // Serve
  log.Printf("starting server on %s", listen)
  log.Fatal(http.ListenAndServe(listen, router))
}

func ListenForEvents(listener *pq.Listener) {
  for {
    select {
      case n := <-listener.Notify:
        if n.Channel == "member_new" {
          redisClient.Set(n.Extra, true, 0)
        } else if n.Channel == "member_delete" {
          redisClient.Del(n.Extra)
        }
      case <- time.After(90 * time.Second):
        go func() {
          listener.Ping()
        }()
    }
  }
}

func GetPermission(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
  userID := p.ByName("userid")
  conversationID := p.ByName("conversationid")

  if userID == "" || conversationID == "" {
    http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
    return
  }

  exists, err := redisClient.Exists(userID + "+" + conversationID).Result()
  if err != nil {
    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
    return
  } else if exists == 0 {
    http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
    return
  }

  w.WriteHeader(200)
}
