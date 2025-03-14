package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"strings"
)

//Козлов Евгений

var (
	rdb *redis.Client
	ctx = context.Background()
)

// Для хранения данных я узнал, что лучше использовать NoSql, я выбрал Redis
func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to redis...")

}

func main() {

	fmt.Println("Starting application")

	http.HandleFunc("/a/", shortLinkHandle)
	http.HandleFunc("/s/", handleRedirect)

	fmt.Println("Server listening port :3001...")
	log.Fatal(http.ListenAndServe(":3001", nil))

}

// handle функция для получения url, который надо сократить
func shortLinkHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//получаю url key
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}
	//Если не указан префикс http/https
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url

	}
	shortCode := generateShortCode(url)

	err := rdb.Set(ctx, shortCode, url, 0).Err()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, shortCode)

}
func generateShortCode(url string) string {
	//добавляем переменную, которая исключит коллизии, т.к. я срезаю код sha256
	salt := 0
	for {
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s%d", url, salt)))

		code := base64.URLEncoding.EncodeToString(hash[:])

		if len(url) > 8 {
			code = code[:8]
		}
		existingUrl, err := rdb.Get(ctx, code).Result()

		if errors.Is(err, redis.Nil) || existingUrl == url {
			return code
		}

		salt++
	}
}

// хэндлер для обработки редиректа
func handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/s/")

	url, err := rdb.Get(ctx, shortCode).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}
