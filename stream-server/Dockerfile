FROM golang:1.23.3
RUN apt-get update && apt-get install -y ffmpeg
WORKDIR /app
COPY . .
RUN go get github.com/go-redis/redis/v8 github.com/prometheus/client_golang/prometheus github.com/prometheus/client_golang/prometheus/promhttp
RUN go build -o stream-server
CMD ["./stream-server"]