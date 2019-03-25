CREATE EXTENSION IF NOT EXISTS CITEXT;

DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE users
(
	id BIGSERIAL CONSTRAINT users_pk PRIMARY KEY,
	nickname CITEXT NOT NULL UNIQUE
	  CONSTRAINT users_nickname_check CHECK ( nickname ~ '^[a-zA-Z0-9_.]+$' ),
	fullname VARCHAR NOT NULL
		CONSTRAINT users_fullname_check CHECK ( fullname <> '' ),
	about TEXT,
	email CITEXT NOT NULL UNIQUE
    CONSTRAINT users_email_check
      CHECK ( email ~ '^[a-zA-Z0-9.!#$%&''*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$' )
);

DROP TABLE IF EXISTS forums CASCADE;
CREATE TABLE forums
(
	id BIGSERIAL CONSTRAINT forums_pk PRIMARY KEY,
	slug CITEXT NOT NULL UNIQUE
	  CONSTRAINT forums_slug_check CHECK ( slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$' ),
	title TEXT NOT NULL
		CONSTRAINT forums_title_check CHECK ( title <> '' ),
	posts BIGINT NOT NULL DEFAULT 0,
	threads INTEGER NOT NULL DEFAULT 0,
	owner CITEXT NOT NULL CONSTRAINT owner_users_fk REFERENCES users (nickname) ON DELETE CASCADE
);

DROP TABLE IF EXISTS threads CASCADE;
CREATE TABLE threads
(
  id BIGSERIAL CONSTRAINT threads_pk PRIMARY KEY,
  slug CITEXT UNIQUE
	  CONSTRAINT threads_slug_check CHECK ( slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$' ),
	title TEXT NOT NULL
    CONSTRAINT threads_title_check CHECK ( title <> '' ),
  message TEXT NOT NULL
    CONSTRAINT threads_message_check CHECK ( message <> '' ),
  votes INTEGER DEFAULT 0,
  created TIMESTAMP WITH TIME ZONE,
  author CITEXT NOT NULL CONSTRAINT author_users_fk REFERENCES users (nickname) ON DELETE CASCADE,
  forum  CITEXT NOT NULL CONSTRAINT parent_forum_fk REFERENCES forums (slug) ON DELETE CASCADE,

  CONSTRAINT uniq_thread UNIQUE (slug, author, forum)
);

DROP TABLE IF EXISTS posts CASCADE;
CREATE TABLE posts
(
  id BIGSERIAL CONSTRAINT posts_pk PRIMARY KEY,
  message TEXT NOT NULL
		CONSTRAINT threads_message_check CHECK ( message <> '' ),
	is_edited BOOLEAN NOT NULL DEFAULT false,
	created TIMESTAMP WITH TIME ZONE NOT NULL,
	author CITEXT NOT NULL CONSTRAINT author_users_fk REFERENCES users (nickname) ON DELETE CASCADE,
	forum  CITEXT NOT NULL CONSTRAINT parent_forum_fk REFERENCES forums (slug) ON DELETE CASCADE,
	thread BIGSERIAL NOT NULL CONSTRAINT thread_fk REFERENCES threads (id) ON DELETE CASCADE,
	parent BIGINT CONSTRAINT parent_post_fk REFERENCES posts(id) ON DELETE CASCADE,
	parents BIGINT[] NOT NULL
);

DROP TABLE IF EXISTS votes CASCADE;
CREATE TABLE votes
(
  author CITEXT NOT NULL CONSTRAINT author_users_fk REFERENCES users (nickname) ON DELETE CASCADE,
  thread BIGSERIAL NOT NULL CONSTRAINT thread_fk REFERENCES threads (id) ON DELETE CASCADE,
  is_up BOOLEAN NOT NULL,

  CONSTRAINT vote_pk PRIMARY KEY (author, thread)
);

DROP FUNCTION IF EXISTS thread_count_increment CASCADE;
CREATE FUNCTION thread_count_increment() RETURNS TRIGGER AS $_$
BEGIN
UPDATE forums SET threads = threads + 1 WHERE slug = new.forum;
RETURN NEW;
END $_$ LANGUAGE 'plpgsql';

CREATE TRIGGER thread_insert_trigger AFTER INSERT ON threads
  FOR EACH ROW EXECUTE PROCEDURE thread_count_increment();

DROP FUNCTION IF EXISTS post_count_increment CASCADE;
CREATE FUNCTION post_count_increment() RETURNS TRIGGER AS $_$
BEGIN
UPDATE forums SET posts = posts + 1 WHERE slug = new.forum;
RETURN NEW;
END $_$ LANGUAGE 'plpgsql';

CREATE TRIGGER post_insert_trigger AFTER INSERT ON posts
  FOR EACH ROW EXECUTE PROCEDURE post_count_increment();