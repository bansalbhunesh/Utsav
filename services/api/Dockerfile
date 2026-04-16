FROM golang:1.25-alpine AS build
WORKDIR /app
COPY services/api/go.mod services/api/go.sum ./services/api/
WORKDIR /app/services/api
RUN go mod download
WORKDIR /app
COPY . .
WORKDIR /app/services/api
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
