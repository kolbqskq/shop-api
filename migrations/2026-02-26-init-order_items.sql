--CreateTable
CREATE TABLE order_items (
    order_id UUID NOT NULL,
    product_id UUID NOT NULL,
    name VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    price BIGINT NOT NULL CHECK(price >= 0),

    PRIMARY KEY (order_id, product_id)

    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);