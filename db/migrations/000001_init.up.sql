BEGIN;

-- Главная таблица заказов
CREATE TABLE IF NOT EXISTS orders (
    order_uid          TEXT PRIMARY KEY,
    track_number       TEXT NOT NULL,
    entry              TEXT,
    locale             TEXT,
    internal_signature TEXT,
    customer_id        TEXT NOT NULL,
    delivery_service   TEXT,
    shardkey           TEXT,
    sm_id              INT,
    date_created       TIMESTAMPTZ NOT NULL,
    oof_shard          TEXT
);

-- Доставка (1:1 с orders)
CREATE TABLE IF NOT EXISTS deliveries (
    order_uid TEXT PRIMARY KEY,
    name      TEXT NOT NULL,
    phone     TEXT,
    zip       TEXT,
    city      TEXT,
    address   TEXT,
    region    TEXT,
    email     TEXT,
    CONSTRAINT deliveries_order_fk
      FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
);

-- Оплата (1:1 с orders)
CREATE TABLE IF NOT EXISTS payments (
    transaction   TEXT PRIMARY KEY,
    order_uid     TEXT UNIQUE,
    request_id    TEXT,
    currency      TEXT NOT NULL,
    provider      TEXT,
    amount        INT  NOT NULL,
    payment_dt    BIGINT NOT NULL,              
    bank          TEXT,
    delivery_cost INT  NOT NULL DEFAULT 0,
    goods_total   INT  NOT NULL DEFAULT 0,
    custom_fee    INT  NOT NULL DEFAULT 0,
    CONSTRAINT payments_order_fk
      FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE,
    CONSTRAINT amount_nonneg CHECK (amount >= 0),
    CONSTRAINT costs_nonneg  CHECK (delivery_cost >= 0 AND goods_total >= 0 AND custom_fee >= 0)
);

-- Товары (N:1 к orders)
CREATE TABLE IF NOT EXISTS items (
    id           BIGSERIAL PRIMARY KEY,        
    order_uid    TEXT NOT NULL,
    chrt_id      INT,
    track_number TEXT,
    price        INT NOT NULL,
    rid          TEXT,
    name         TEXT,
    sale         INT,
    size         TEXT,
    total_price  INT NOT NULL,
    nm_id        INT,
    brand        TEXT,
    status       INT,
    CONSTRAINT items_order_fk
      FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE,
    CONSTRAINT price_nonneg CHECK (price >= 0 AND total_price >= 0)
);

CREATE INDEX IF NOT EXISTS idx_items_order_uid     ON items(order_uid);
CREATE INDEX IF NOT EXISTS idx_payments_order_uid  ON payments(order_uid);
CREATE INDEX IF NOT EXISTS idx_orders_date         ON orders(date_created);

COMMIT;
