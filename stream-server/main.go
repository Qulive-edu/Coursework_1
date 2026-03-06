package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type videos struct {
	VideoArr []string `json:"videos"`
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func listVideosHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Запрос на /videos пришел")
	files, err := os.ReadDir("./videos")
	if err != nil {
		http.Error(w, "Ошибка чтения директории", http.StatusInternalServerError)
		return
	}
	var videoFiles videos
	fmt.Println(videoFiles)
	for _, file := range files {
		if !file.IsDir() && (filepath.Ext(file.Name()) == ".mp4" || filepath.Ext(file.Name()) == ".mkv") {
			videoFiles.VideoArr = append(videoFiles.VideoArr, file.Name())
			fmt.Println(videoFiles.VideoArr)
			fmt.Printf("Видеофайл с именем %s был найден и отправлен", file.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Println(videoFiles.VideoArr)
	fmt.Println("Теперь json")
	js, err := json.Marshal(videoFiles)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Println(string(js))
	json.NewEncoder(w).Encode(videoFiles)
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
	fmt.Println("Получен запрос на потоковую передачу")
	w.Header().Set("Content-Type", "video/mp4")

	cacheKey := "current_stream"
	if cachedData, found := getCachedStreamData(cacheKey); found {
		fmt.Println("Отправляем с кэша")
		w.Write(cachedData)
		return
	}

	file := r.URL.Query().Get("file")
	if file == "" {
		fmt.Println("Файл не найден, нет имени")
		http.Error(w, "Не указано имя файла", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join("./videos", file)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}
	fmt.Printf("Путь до файла: %s", filePath)

	cmd := exec.Command("ffmpeg", "-re", "-i", filePath, "-f", "mp4", "-movflags", "frag_keyframe+empty_moov", "pipe:1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, "Ошибка запуска ffmpeg", http.StatusInternalServerError)
		return
	}

	if err := cmd.Start(); err != nil {
		http.Error(w, "Ошибка выполнения ffmpeg", http.StatusInternalServerError)
		return
	}

	dataBuffer := make([]byte, 1024)
	fmt.Println("Начинается передача видео")
	for {
		n, err := stdout.Read(dataBuffer)
		if n > 0 {
			w.Write(dataBuffer[:n])
			cacheStreamData(cacheKey, dataBuffer[:n])
			w.(http.Flusher).Flush()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Ошибка чтения из ffmpeg:", err)
			break
		}
	}

	cmd.Wait()
}

func main() {
	initRedis()

	mux := http.NewServeMux()

	mux.HandleFunc("/videos", listVideosHandler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/stream", streamHandler)
	mux.HandleFunc("/upload", uploadVideoHandler)

	fmt.Println("Потоковый сервер запущен на порту 8080")

	if err := http.ListenAndServe(":8080", corsMiddleware(mux)); err != nil {
		fmt.Println("Ошибка при запуске сервера:", err)
	}
}
