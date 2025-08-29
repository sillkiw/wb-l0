-- Главная таблица заказов
CREATE TABLE IF NOT EXISTS orders (
    order_uid          TEXT PRIMARY KEY,
    track_number       TEXT,
    entry              TEXT,
    locale             TEXT,
    internal_signature TEXT,
    customer_id        TEXT,
    delivery_service   TEXT,
    shardkey           TEXT,
    sm_id              INT,
    date_created       TIMESTAMP,
    oof_shard          TEXT
);

-- Доставка (одна запись на заказ)
CREATE TABLE IF NOT EXISTS deliveries (
    order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name      TEXT,
    phone     TEXT,
    zip       TEXT,
    city      TEXT,
    address   TEXT,
    region    TEXT,
    email     TEXT
);

-- Оплата (одна запись на заказ)
CREATE TABLE IF NOT EXISTS payments (
    transaction   TEXT PRIMARY KEY,
    order_uid     TEXT UNIQUE REFERENCES orders(order_uid) ON DELETE CASCADE,
    request_id    TEXT,
    currency      TEXT,
    provider      TEXT,
    amount        INT,
    payment_dt    BIGINT,
    bank          TEXT,
    delivery_cost INT,
    goods_total   INT,
    custom_fee    INT
);

-- Товары (много на заказ)
CREATE TABLE IF NOT EXISTS items (
    id           SERIAL PRIMARY KEY,
    order_uid    TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id      INT,
    track_number TEXT,
    price        INT,
    rid          TEXT,
    name         TEXT,
    sale         INT,
    size         TEXT,
    total_price  INT,
    nm_id        INT,
    brand        TEXT,
    status       INT
);
