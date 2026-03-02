--CreateTable
CREATE TABLE products (
    id UUID PRIMARY KEY,
    name VARCHAR(20) NOT NULL,
    description VARCHAR(200) NOT NULL,
    category VARCHAR(20) NOT NULL,
    price BIGINT NOT NULL CHECK (price >= 0),
    stock INT NOT NULL CHECK (stock >= 0),
    reserved INT NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    is_active BOOLEAN NOT NULL,
    version BIGINT NOT NULL DEFAULT 0 CHECK (version >= 0),
    CHECK (reserved <= stock),
    created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
