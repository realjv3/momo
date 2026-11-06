.PHONY: run replay down

DATE ?= 2026-11-03
SYMBOLS ?= PARA,CHSN
PORT ?= 7778

NODE_BIN := $(shell ls -d $(HOME)/.nvm/versions/node/*/bin 2>/dev/null | sort -V | tail -1)
export PATH := $(NODE_BIN):$(PATH)

run:
	docker compose up -d
	cd web && npm install && npm run build
	go run ./cmd/momo serve

replay:
	go run ./cmd/momo replay -date $(DATE) -symbols $(SYMBOLS)

down:
	@pids=$$(lsof -ti tcp:$(PORT) -sTCP:LISTEN); [ -z "$$pids" ] || kill $$pids
	docker compose down
