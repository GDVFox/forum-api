DROP TABLE IF EXISTS posts;
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

	CONSTRAINT uniq_post UNIQUE (author, forum)
);