CREATE TABLE IF NOT EXISTS users
(
    "id"                SERIAL,
    "tg_id"             bigint primary key,
    "bal"               double precision default 0,
    "capital"           double precision default 0,
    "percent"           double precision,
    "api_key"           text             default '',
    "secret_key"        text             default '',
    "trading_activated" boolean          default false,
    "status"            text             default 'none'
);

CREATE TABLE IF NOT EXISTS coin
(
    "user_id"       bigint references users (tg_id),
    "coin_name"     text,
    "entry_price"   double precision default 0,
    "decrement"     double precision default 0,
    "count"         double precision default 0,
    "buy"           double precision[],
    "buy_order_id"  text             default '',
    "sell_order_id" text             default '',
    unique (coin_name, user_id)
);

CREATE TABLE IF NOT EXISTS income
(
    user_id     bigint references users (tg_id),
    "coin_name" text,
    "count"     double precision default 0,
    "income"    double precision default 0,
    "time"      timestamp        default now() not null
);
CREATE TABLE IF NOT EXISTS coiniks
(
    "coin_name"      text,
    "qty_decimals"   int              default 0,
    "price_decimals" int              default 0,
    "min_sum_buy"    double precision default 0
);