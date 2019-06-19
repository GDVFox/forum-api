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

DROP FUNCTION IF EXISTS post_count_increment CASCADE;
CREATE FUNCTION post_count_increment() RETURNS TRIGGER AS $_$
BEGIN
UPDATE forums SET posts = posts + 1 WHERE slug = new.forum;
RETURN NEW;
END $_$ LANGUAGE 'plpgsql';

CREATE TRIGGER post_insert_trigger AFTER INSERT ON posts
  FOR EACH ROW EXECUTE PROCEDURE post_count_increment();

CREATE INDEX posts_thread_index
  ON posts (thread);

CREATE INDEX posts_thread_id_index
  ON posts (thread, id);

CREATE INDEX ON posts (thread, id, parent)
  WHERE parent IS NULL;

CREATE INDEX parent_tree
  ON posts (parents DESC, id);