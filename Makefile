# =========================
# APP INFO
# =========================
APP_NAME=main
MAIN_FILE=./cmd/api

# =========================
# GO COMMANDS
# =========================
run:
	go run $(MAIN_FILE)

tidy:
	go mod tidy

download:
	go mod download

# =========================
# TEST
# =========================
test:
	go test -v ./...

test-all:
	go test -v \
	./internal/tests/unit/services/... \
	./internal/tests/unit/repositories/... \
	./internal/tests/unit/handlers/... \
	./internal/tests/integration/... \
	--coverpkg=./internal/services/...,./internal/repositories/...,./internal/handlers/...

test-unit-cover:
	go test -v ./internal/tests/unit/... --coverpkg=./internal/...

test-unit-service:
	go test -v ./internal/tests/unit/services/... --coverpkg=./internal/services/...

test-unit-handler:
	go test -v ./internal/tests/unit/handlers/... --coverpkg=./internal/handlers/...

test-integration:
	go test -v ./internal/tests/integration/... --coverpkg=./internal/...




# =========================
# DOCKER
# =========================
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
