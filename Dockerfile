FROM golang:1.26-alpine3.23 AS modules
WORKDIR /modules
COPY go.mod go.sum ./
RUN go mod download

FROM golang:1.26-alpine3.23 AS builder
WORKDIR /app

COPY --from=modules /go/pkg /go/pkg
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -tags service -o /bin/app_service ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -tags migrate -o /bin/app_migrate ./cmd/migrate

FROM scratch AS production

WORKDIR /
COPY --from=builder /app/configs /config
COPY --from=builder /app/db/migrations /db/migrations
COPY --from=builder /app/.env.docker /.env.docker

COPY --from=builder /bin/app_service /app_service
COPY --from=builder /bin/app_migrate /app_migrate

EXPOSE 3000
CMD ["/app_service"]

FROM golang:1.26-alpine3.23 AS dev

WORKDIR /app

COPY --from=modules /go/pkg /go/pkg

RUN go install github.com/cespare/reflex@latest

COPY . .

EXPOSE 3000

CMD ["go", "run", "./cmd/api"]