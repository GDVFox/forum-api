DROP TABLE IF EXISTS forums;
CREATE TABLE forums
(
	id BIGSERIAL CONSTRAINT forums_pk PRIMARY KEY,
	slug CITEXT COLLATE "en_US.utf8" NOT NULL UNIQUE
	  CONSTRAINT forums_slug_check CHECK ( slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$' ),
	title TEXT NOT NULL
		CONSTRAINT users_title_check CHECK ( title <> '' ),
	posts BIGINT NOT NULL DEFAULT 0,
	threads INTEGER NOT NULL DEFAULT 0,
	owner CITEXT NOT NULL CONSTRAINT owner_users_fk REFERENCES users (nickname) ON DELETE CASCADE
);

CREATE INDEX forum_cover_index
  ON forums (id, slug, title, threads, posts, owner);