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
  created TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
  author CITEXT NOT NULL CONSTRAINT author_users_fk REFERENCES users (nickname) ON DELETE CASCADE,
  forum  CITEXT NOT NULL CONSTRAINT parent_forum_fk REFERENCES forums (slug) ON DELETE CASCADE,

  CONSTRAINT uniq_thread UNIQUE (slug, author, forum)
);