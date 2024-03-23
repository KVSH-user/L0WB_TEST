-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR PRIMARY KEY,
    track_number VARCHAR,
    entry VARCHAR,
    locale VARCHAR,
    internal_signature VARCHAR,
    customer_id VARCHAR,
    delivery_service VARCHAR,
    shardkey VARCHAR,
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard VARCHAR
);

CREATE TABLE IF NOT EXISTS delivery (
    order_uid VARCHAR PRIMARY KEY,
    name VARCHAR,
    phone VARCHAR,
    zip VARCHAR,
    city VARCHAR,
    address VARCHAR,
    region VARCHAR,
    email VARCHAR,
    FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS payment (
    order_uid VARCHAR PRIMARY KEY,
    transaction VARCHAR,
    request_id VARCHAR,
    currency VARCHAR,
    provider VARCHAR,
    amount INT,
    payment_dt TIMESTAMP,
    bank VARCHAR,
    delivery_cost INT,
    goods_total INT,
    custom_fee INT,
    FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS items (
    item_id SERIAL PRIMARY KEY,
    order_uid VARCHAR,
    chrt_id INT,
    track_number VARCHAR,
    price INT,
    rid VARCHAR,
    name VARCHAR,
    sale INT,
    size VARCHAR,
    total_price INT,
    nm_id INT,
    brand VARCHAR,
    status INT,
    FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS payment;
DROP TABLE IF EXISTS delivery;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
