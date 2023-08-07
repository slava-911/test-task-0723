CREATE TABLE users(
    id         UUID PRIMARY KEY,
    firstname  TEXT NOT NULL,
    lastname   TEXT NOT NULL,
    email      TEXT NOT NULL,
    password   TEXT NOT NULL,
    age        INT NOT NULL,
    is_married BOOLEAN
);

CREATE TABLE products(
    id          UUID PRIMARY KEY,
    price       BIGINT NOT NULL,
    quantity    BIGINT NOT NULL,
    description TEXT   NOT NULL,
    tags        TEXT[]
);

CREATE TABLE orders(
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL,
    created_at TIMESTAMP NOT NULL,
    completed  BOOLEAN
);

CREATE TABLE orders_content(
    order_id   UUID   NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id UUID   NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    price      BIGINT NOT NULL,
    quantity   BIGINT NOT NULL,
    PRIMARY KEY (order_id, product_id)
);

CREATE TABLE orders_content_history(
    operation CHAR(1) NOT NULL,
    stamp TIMESTAMP NOT NULL,
    order_id   UUID   NOT NULL,
    product_id UUID   NOT NULL,
    price      BIGINT NOT NULL,
    quantity   BIGINT NOT NULL,
    PRIMARY KEY (stamp, order_id, product_id)
);

CREATE OR REPLACE FUNCTION process_orders_content_history() RETURNS TRIGGER AS $orders_content_history$
    BEGIN
        IF (TG_OP = 'DELETE') THEN
            INSERT INTO orders_content_history SELECT 'D', now(), OLD.*;
        ELSIF (TG_OP = 'UPDATE') THEN
            INSERT INTO orders_content_history SELECT 'U', now(), NEW.*;
        ELSIF (TG_OP = 'INSERT') THEN
            INSERT INTO orders_content_history SELECT 'I', now(), NEW.*;
        END IF;
        RETURN NULL;
    END;
$orders_content_history$ LANGUAGE plpgsql;

CREATE TRIGGER orders_content_history
AFTER INSERT OR UPDATE OR DELETE ON orders_content
    FOR EACH ROW EXECUTE FUNCTION process_orders_content_history();