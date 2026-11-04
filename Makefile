.PHONY: run

run:
	docker compose up -d
	cd web && npm install && npm run build
	go run ./cmd/momo serve
