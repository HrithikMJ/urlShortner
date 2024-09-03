package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var redisCtx = context.Background()
var pgCtx = context.Background()
var rdb *redis.Client

func secureRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return ""
		}
		result[i] = letters[num.Int64()]
	}

	return string(result)

}
func getHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(secureRandomString(6))
	w.Write([]byte("Hi"))
}
func setUrlHandler(w http.ResponseWriter, r *http.Request) {
	url := r.PathValue("url")
	// shortUrl, err := rdb.Get(redisCtx, url).Result()
	// if err != nil {
	val := secureRandomString(6)
	err := rdb.Set(redisCtx, val, url, 3600*time.Second).Err()
	if err != nil {
	}
	w.Write([]byte(val))
	// } else {
	// 	w.Write([]byte(fmt.Sprintf("Url is already shortned %v", shortUrl)))
	// }
}
func getIdHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	url, err := rdb.Get(redisCtx, id).Result()
	fmt.Println(url)
	formattedUrl := fmt.Sprintf("https://" + url)
	if err == nil {
		http.Redirect(w, r, formattedUrl, http.StatusSeeOther)

	} else {
		w.Write([]byte(fmt.Sprintf("%v is not in db!", r.PathValue("id"))))
	}
}
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	err := rdb.FlushDBAsync(context.Background()).Err()
	if err != nil {
		fmt.Println("Error flushing database:", err)
	}
	fmt.Println("Flushed database!")
	w.Write([]byte("Flushed database!"))

}
func main() {
	mux := http.NewServeMux()
	_, err := pgxpool.New(pgCtx, "postgres:///urldb?host=/var/lib/postgresql")
	if err != nil {
		fmt.Println(err)
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	fmt.Println("Url Shortner", rdb)
	mux.HandleFunc("GET /{id}", getIdHandler)
	mux.HandleFunc("GET /url/{url}", setUrlHandler)
	mux.HandleFunc("GET /", getHandler)
	mux.HandleFunc("GET /delete", deleteHandler)
	srv := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
	srv.ListenAndServe()
}
