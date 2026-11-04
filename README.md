# momo

NASDAQ small-cap momentum scanner. Go backend, Postgres, React UI.

## Run

```sh
docker compose up -d                  # postgres on :5433
cp .env.example .env                  # set MASSIVE_API_KEY
go run ./cmd/momo serve               # API + websocket on :7778 (first run loads the universe)
cd web && npm install && npm run dev  # UI on :5173, proxies /api to :7778
```

`cd web && npm run build` makes `web/dist`, which `serve` hosts at `:7778`.

## Commands

- `serve` live scanner; needs a plan with websocket access (`MASSIVE_FEED=delayed|realtime`)
- `refresh` reload NASDAQ universe, free float and previous closes
- `replay -date 2026-10-30 -symbols CHOW,FRGT` run historical minute bars through the engine

Set `MOMO_DEV=1` to enable `POST /api/dev/signal` for injecting a test row.
