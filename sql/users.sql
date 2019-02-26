DROP TABLE IF EXISTS users
CREATE TABLE users
(
	id BIGSERIAL CONSTRAINT users_pk PRIMARY KEY,
	nickname CITEXT COLLATE "en_US.utf8" NOT NULL UNIQUE
	  CONSTRAINT users_nickname_check CHECK ( nickname ~ '^[a-zA-Z0-9_]+$' ),
	fullname VARCHAR NOT NULL 
		CONSTRAINT users_fullname_check CHECK ( fullname <> '' ),
	about TEXT,
	email CITEXT COLLATE "en_US.utf8" NOT NULL UNIQUE
    CONSTRAINT users_email_check
      CHECK ( email ~ '^[a-zA-Z0-9.!#$%&''*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$' )
);

CREATE unique index users_nickname_uindex
  ON users (nickname);