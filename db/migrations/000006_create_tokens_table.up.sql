-- create tokens table
-- delete record if user being referenced gets deleted
-- user_id references primary key of users table
CREATE TABLE IF NOT EXISTS tokens (
  hash bytea PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  expiry timestamp(0) with time zone NOT NULL,
  scope text NOT NULL
);
