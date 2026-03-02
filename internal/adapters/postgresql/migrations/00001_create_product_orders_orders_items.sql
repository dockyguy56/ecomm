-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS products (
  id int PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  image varchar(255) NOT NULL,
  category varchar(255) NOT NULL,
  description text,
  rating int NOT NULL,
  num_reviews int NOT NULL DEFAULT 0,
  price decimal(10,2) NOT NULL,
  count_in_stock int NOT NULL,
  created_at TIMESTAMPTZ DEFAULT (now()),
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS orders (
  id int PRIMARY KEY NOT NULL,
  payment_method varchar(255) NOT NULL,
  tax_price decimal(10,2) NOT NULL,
  shipping_price decimal(10,2) NOT NULL,
  total_price decimal(10,2) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT (now()),
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS order_items (
  id int PRIMARY KEY NOT NULL,
  order_id int NOT NULL,
  product_id int NOT NULL,
  name varchar(255) NOT NULL,
  quantity int NOT NULL,
  image varchar(255) NOT NULL,
  price int NOT NULL
);

ALTER TABLE order_items ADD FOREIGN KEY (order_id) REFERENCES orders (id);

ALTER TABLE order_items ADD FOREIGN KEY (product_id) REFERENCES products (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
