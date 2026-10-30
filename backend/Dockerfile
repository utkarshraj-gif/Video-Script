FROM golang:1.27.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o video-rag-backend .

FROM python:3.12-alpine

WORKDIR /app

RUN apk add --no-cache ffmpeg ca-certificates nodejs
RUN pip install --no-cache-dir "yt-dlp[default]"

# Copy the compiled Go application
COPY --from=builder /app/video-rag-backend .

EXPOSE 8080

CMD ["./video-rag-backend"]