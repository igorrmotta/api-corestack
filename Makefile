.PHONY: db-up db-down db-migrate db-rollback db-new \
       proto-gen proto-lint proto-breaking \
       go-server go-worker go-test \
       test-bruno docker-go

# ── Database ──────────────────────────────────────────────

db-up:
	docker compose up postgres -d

db-down:
	docker compose down

db-migrate:
	docker compose run --rm migrate up

db-rollback:
	docker compose run --rm migrate down

db-new:
	@test -n "$(NAME)" || (echo "Usage: make db-new NAME=create_tasks" && exit 1)
	docker compose run --rm migrate new $(NAME)

# ── Protobuf ─────────────────────────────────────────────

proto-gen:
	cd api && buf generate

proto-lint:
	cd api && buf lint

proto-breaking:
	cd api && buf breaking --against '.git#branch=main'

# ── Go ───────────────────────────────────────────────────

go-server:
	cd services/golang && go run ./cmd/server

go-worker:
	cd services/golang && go run ./cmd/worker

go-test:
	cd services/golang && go test ./...

# ── Integration Tests ────────────────────────────────────

test-bruno:
	npx @usebruno/cli run tests/bruno --env go-dev

# ── Docker ───────────────────────────────────────────────

docker-go:
	docker compose --profile go up --build
