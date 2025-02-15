BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    coins INTEGER NOT NULL DEFAULT 1000
);

CREATE TABLE IF NOT EXISTS merch_items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    price INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS user_inventory (
    user_id INTEGER REFERENCES users(id),
    item_id INTEGER REFERENCES merch_items(id),
    quantity INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, item_id)
);

CREATE TABLE IF NOT EXISTS coin_transactions (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER REFERENCES users(id),
    receiver_id INTEGER REFERENCES users(id),
    amount INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO merch_items (name, price)
VALUES 
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500)
ON CONFLICT (name) DO NOTHING;


COMMIT;