# API image only. Build context: services/api (Render "Docker Build Context" = services/api).
# Local: cd services/api && docker build -f Dockerfile .
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV GOMAXPROCS=1
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/utsav-api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 10001 utsav
USER utsav
WORKDIR /app
COPY --from=build /bin/utsav-api /app/utsav-api
EXPOSE 8080
CMD ["/app/utsav-api"]
