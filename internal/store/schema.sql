create table if not exists symbols (
    symbol          text primary key,
    free_float      bigint not null default 0,
    prev_close      double precision not null default 0,
    prev_close_date date,
    updated_at      timestamptz not null default now()
);

create table if not exists signals (
    id         bigserial primary key,
    ts         timestamptz not null,
    symbol     text not null,
    session    text not null,
    change_pct double precision not null,
    price      double precision not null,
    free_float bigint not null,
    volume     double precision not null
);

create index if not exists signals_ts_idx on signals (ts desc);

create table if not exists settings (
    key   text primary key,
    value jsonb not null
);

alter table signals drop column if exists bar_volume, drop column if exists ema;
alter table symbols drop column if exists name, drop column if exists free_float_pct;
