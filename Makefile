.PHONY: run replay

DATE ?= 2026-11-03
SYMBOLS ?= PARA,CHSN

run:
	docker compose up -d
	cd web && npm install && npm run build
	go run ./cmd/momo serve

replay:
	go run ./cmd/momo replay -date $(DATE) -symbols $(SYMBOLS)
