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

type VideoMeta struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	Ext     string    `json:"ext"`
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
	cacheHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "video_cache_hits_total",
			Help: "Number of cache hits for video metadata",
		},
	)
	cacheMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "video_cache_misses_total",
			Help: "Number of cache misses for video metadata",
		},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal, cacheHits, cacheMisses)
}

func initRedis() error {
	redisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("не удалось подключиться к Redis: %w", err)
	}
	fmt.Println("Подключено к Redis")
	return nil
}

func cacheVideoMeta(key string, meta VideoMeta) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(meta)
	if err != nil {
		fmt.Printf("Ошибка маршалинга метаданных: %v\n", err)
		return
	}

	if err := redisClient.Set(ctx, key, data, 10*time.Minute).Err(); err != nil {
		fmt.Printf("Ошибка записи в Redis: %v\n", err)
	}
}

func getCachedVideoMeta(key string) (*VideoMeta, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		fmt.Printf("Ошибка чтения из Redis: %v\n", err)
		return nil, false
	}

	var meta VideoMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		fmt.Printf("Ошибка размаршалинга метаданных: %v\n", err)
		return nil, false
	}
	return &meta, true
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	requestsTotal.Inc()

	file := r.URL.Query().Get("file")
	if file == "" {
		http.Error(w, "Не указано имя файла", http.StatusBadRequest)
		return
	}

	cacheKey := "video:meta:" + file

	if meta, found := getCachedVideoMeta(cacheKey); found {
		cacheHits.Inc()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		if err := json.NewEncoder(w).Encode(meta); err != nil {
			fmt.Printf("Ошибка отправки кэша: %v\n", err)
		}
		fmt.Printf("Метаданные отданы из кэша: %s\n", file)
		return
	}

	cacheMisses.Inc()

	videoDir := os.Getenv("VIDEO_DIRECTORY")
	if videoDir == "" {
		videoDir = "./videos"
	}
	filePath := filepath.Join(videoDir, file)

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		fmt.Printf("Файл не найден: %s\n", filePath)
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		fmt.Printf("Ошибка stat файла: %v\n", err)
		http.Error(w, "Ошибка доступа к файлу", http.StatusInternalServerError)
		return
	}

	meta := VideoMeta{
		Name:    file,
		Size:    info.Size(),
		ModTime: info.ModTime(),
		Ext:     filepath.Ext(file),
	}

	go cacheVideoMeta(cacheKey, meta)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		fmt.Printf("Ошибка отправки метаданных: %v\n", err)
	}
	fmt.Printf("Метаданные прочитаны с диска и закэшированы: %s\n", file)
}

func main() {
	if err := initRedis(); err != nil {
		fmt.Printf("Redis недоступен: %v (продолжаем без кэша)\n", err)
	}

	http.HandleFunc("/videos", listVideosHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/stream", streamHandler)
	http.HandleFunc("/upload", uploadVideoHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Потоковый сервер запущен на порту %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %v\n", err)
	}
}
