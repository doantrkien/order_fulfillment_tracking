# =========================
# APP INFO
# =========================
APP_NAME=main
MAIN_FILE=./cmd/api
MIGRATION_FILE=./cmd/migrate

# =========================
# GO COMMANDS
# =========================
run: export ENV_FILE=.env.local
run:
	go run $(MAIN_FILE)

tidy:
	go mod tidy

download:
	go mod download

migrate: export ENV_FILE=.env.local
migrate:
	go run $(MIGRATION_FILE)

migrate-up:
	docker exec -it order_tracking go run ./cmd/migrate

seed:
	docker exec -i postgres psql -U postgres -d order_tracking < db/seed.sql

# =========================
# SWAGGER
# =========================
swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g $(MAIN_FILE)/main.go

# =========================
# TEST
# =========================
test:
	go test -v ./...

test-all:
	go test -v \
	./internal/tests/unit/services/... \
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

docker-up-infra:
	docker compose up -d postgres adminer 

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-build:
	docker compose up -d --build




# =========================
# OLLAMA SETUP
# =========================
setup-ollama:
	docker compose up -d ollama
	docker compose exec -T ollama sh -c "until ollama list >/dev/null 2>&1; do sleep 2; done"
	docker compose exec -T ollama ollama pull nomic-embed-text-v2-moe

