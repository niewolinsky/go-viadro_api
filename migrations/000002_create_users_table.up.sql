CREATE TABLE IF NOT EXISTS users (
    user_id serial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    username text,
    email citext UNIQUE,
    password_hash bytea,
    activated bool
);