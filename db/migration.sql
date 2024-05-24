CREATE TABLE users
(
    "id" SERIAL,
    "tg_id" bigint primary key,
    "bal" double precision,
    "capital" double precision,
    "percent" double precision,
    "income" double precision,
    "status" text
);
CREATE TABLE coin
(
    "user_id" bigint references users (tg_id),
    "coin_name" text,
    "entry_price" double precision,
    "decrement" double precision,
    "count" bigint,
    "buy" double precision[]
);