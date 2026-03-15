--CreateTable
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total BIGINT NOT NULL CHECK(total >= 0),
    created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version BIGINT NOT NULL DEFAULT 0,

    FOREIGN KEY (user_id) REFERENCES users(id)
);