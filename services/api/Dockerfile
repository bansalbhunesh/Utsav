# Build context must be the services/api directory (e.g. Render "Docker Build Context" = services/api).
# Local: cd services/api && docker build -f Dockerfile .
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/utsav-api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/utsav-worker ./cmd/worker

FROM alpine:3.20
RUN adduser -D -u 10001 utsav
USER utsav
WORKDIR /app
COPY --from=build /bin/utsav-api /app/utsav-api
COPY --from=build /bin/utsav-worker /app/utsav-worker
EXPOSE 8080
# Default process is API. Render "Docker Command" overrides CMD (not ENTRYPOINT), so the
# worker Web Service can set Docker Command to: /app/utsav-worker
CMD ["/app/utsav-api"]
