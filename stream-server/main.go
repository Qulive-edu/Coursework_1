package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type videos struct {
	VideoArr []string `json:"videos"`
}

func listVideosHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Запрос на /videos пришел")

	videoDir := os.Getenv("VIDEO_DIRECTORY")
	if videoDir == "" {
		videoDir = "./videos"
	}

	files, err := os.ReadDir(videoDir)
	if err != nil {
		fmt.Printf("Ошибка чтения директории %s: %v\n", videoDir, err)
		http.Error(w, "Ошибка чтения директории: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var videoFiles videos
	for _, file := range files {
		if !file.IsDir() && (filepath.Ext(file.Name()) == ".mp4" || filepath.Ext(file.Name()) == ".mkv") {
			videoFiles.VideoArr = append(videoFiles.VideoArr, file.Name())
			fmt.Printf("Найден видеофайл: %s\n", file.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(videoFiles); err != nil {
		fmt.Printf("Ошибка кодирования JSON: %v\n", err)
	}

	fmt.Printf("Отправлено видео: %v\n", videoFiles.VideoArr)
}

func uploadVideoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	file, handler, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Ошибка загрузки файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dest, err := os.Create(filepath.Join("./videos", handler.Filename))
	if err != nil {
		http.Error(w, "Ошибка сохранения файла", http.StatusInternalServerError)
		return
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		http.Error(w, "Ошибка копирования файла", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Файл %s успешно загружен", handler.Filename)
}

var redisClient *redis.Client

var (
	requestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
}

func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
}

func cacheStreamData(key string, data []byte) {
	ctx := context.Background()
	redisClient.Set(ctx, key, data, 10*time.Minute)
}

func getCachedStreamData(key string) ([]byte, bool) {
	ctx := context.Background()
	data, err := redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false
	}
	return data, true
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	requestsTotal.Inc()

	file := r.URL.Query().Get("file")
	if file == "" {
		http.Error(w, "Не указано имя файла", http.StatusBadRequest)
		return
	}

	videoDir := os.Getenv("VIDEO_DIRECTORY")
	if videoDir == "" {
		videoDir = "./videos"
	}

	filePath := filepath.Join(videoDir, file)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Файл не найден: %s\n", filePath)
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}

	fmt.Printf("Отдаём файл: %s\n", filePath)
	http.ServeFile(w, r, filePath)
}

func main() {
	initRedis()

	http.HandleFunc("/videos", listVideosHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/stream", streamHandler)
	http.HandleFunc("/upload", uploadVideoHandler)
	fmt.Println("Потоковый сервер запущен на порту 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Ошибка при запуске сервера:", err)
	}
}
