CREATE TABLE IF NOT EXISTS "user" (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(30),
    password VARCHAR(30),
    email VARCHAR(320)
);