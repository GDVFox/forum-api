DROP TABLE IF EXISTS threads;
CREATE TABLE threads
(
  id BIGSERIAL CONSTRAINT  threads_pk PRIMARY KEY,
  slug CITEXT COLLATE "en_US.utf8" UNIQUE
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

DROP FUNCTION IF EXISTS thread_count_increment CASCADE;
CREATE FUNCTION thread_count_increment() RETURNS TRIGGER AS $_$
BEGIN
UPDATE forums SET threads = threads + 1 WHERE slug = new.forum;
RETURN NEW;
END $_$ LANGUAGE 'plpgsql';

CREATE TRIGGER thread_insert_trigger AFTER INSERT ON threads
  FOR EACH ROW EXECUTE PROCEDURE thread_count_increment();