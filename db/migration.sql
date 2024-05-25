CREATE TABLE users
(
    "id" SERIAL,
    "tg_id" bigint primary key,
    "bal" double precision default 0,
    "capital" double precision default 0,
    "percent" double precision default 1,
    "income" double precision default 0,
    "status" text
);
CREATE TABLE coin
(
    "user_id" bigint references users (tg_id),
    "coin_name" text,
    "entry_price" double precision default 0,
    "decrement" double precision default 0,
    "count" bigint default 0,
    "buy" double precision[]
);