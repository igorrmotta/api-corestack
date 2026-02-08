.PHONY: db-up db-down db-migrate db-rollback db-new \
       proto-gen proto-lint proto-breaking \
       go-server go-worker go-test \
       ts-install ts-server ts-worker ts-test \
       test-bruno test-bruno-ts docker-go docker-ts

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

# ── TypeScript ───────────────────────────────────────────

ts-install:
	cd services/typescript && pnpm install

ts-server:
	cd services/typescript && pnpm dev:server

ts-worker:
	cd services/typescript && pnpm dev:worker

ts-test:
	cd services/typescript && pnpm test

# ── Integration Tests ────────────────────────────────────

test-bruno-ts:
	npx @usebruno/cli run tests/bruno --env ts-dev

# ── Docker ───────────────────────────────────────────────

docker-go:
	docker compose --profile go up --build

docker-ts:
	docker compose --profile typescript up --build
