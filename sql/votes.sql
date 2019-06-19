DROP TABLE IF EXISTS votes CASCADE;
CREATE TABLE votes
(
  author CITEXT NOT NULL CONSTRAINT author_users_fk REFERENCES users (nickname) ON DELETE CASCADE,
  thread BIGSERIAL NOT NULL CONSTRAINT thread_fk REFERENCES threads (id) ON DELETE CASCADE,
  is_up BOOLEAN NOT NULL,

  CONSTRAINT vote_pk PRIMARY KEY (author, thread)
);