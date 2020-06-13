DROP TABLE IF EXISTS forum_users;
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS threads;
DROP TABLE IF EXISTS forums;
DROP TABLE IF EXISTS users;

CREATE EXTENSION IF NOT EXISTS CITEXT;

CREATE TABLE users
(
  id  SERIAL PRIMARY KEY,
  about TEXT,
  email CITEXT UNIQUE NOT NULL,
  fullname VARCHAR(256) NOT NULL,
  nickname CITEXT UNIQUE NOT NULL
);

CREATE INDEX IF NOT EXISTS users_nickname_and_email ON users (nickname, email);

CREATE TABLE forums
(
  id SERIAL PRIMARY KEY,
  posts BIGINT NOT NULL DEFAULT 0,
  slug CITEXT UNIQUE NOT NULL,
  threads INT NOT NULL DEFAULT 0,
  title TEXT NOT NULL,
  author CITEXT NOT NULL REFERENCES users (nickname)
);

CREATE INDEX IF NOT EXISTS forums_slug ON forums USING hash (slug);

CREATE TABLE forum_users
(
  fUser CITEXT COLLATE ucs_basic NOT NULL,
  forum CITEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS f_users_ind ON forum_users (forum, fUser);
CREATE INDEX  IF NOT EXISTS forum_users_fuser ON forum_users (fUser);

-- TODO add forum's id
CREATE TABLE threads
(
  author CITEXT NOT NULL REFERENCES users (nickname),
  created TIMESTAMP(3) WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  forum CITEXT NOT NULL ,
  id SERIAL PRIMARY KEY,
  message TEXT NOT NULL,
  slug CITEXT UNIQUE,
  title TEXT,
  votes INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS threads_id ON threads (id);
CREATE INDEX IF NOT EXISTS threads_slug_and_id ON threads (slug, id);
CREATE INDEX IF NOT EXISTS threads_created ON threads (created);
CREATE INDEX IF NOT EXISTS threads_forum_created ON threads (forum, created);

CREATE TABLE posts
(
  id SERIAL PRIMARY KEY,
  author CITEXT NOT NULL REFERENCES users (nickname),
  created TIMESTAMP(3) WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  forum CITEXT NOT NULL,
  isEdited BOOLEAN DEFAULT FALSE,
  message TEXT NOT NULL,
  parent INT DEFAULT 0,
  tid INT NOT NULL REFERENCES threads (id),
  slug INTEGER[] NOT NULL,
  rootId INT
);

CREATE INDEX IF NOT EXISTS posts_id ON posts (id);
CREATE INDEX IF NOT EXISTS posts_tid_and_slug ON posts (tid, slug);
CREATE INDEX IF NOT EXISTS posts_tid_and_id ON posts (tid, id);
CREATE INDEX IF NOT EXISTS posts_tid_parent_id ON posts (tid, parent, id);
CREATE INDEX IF NOT EXISTS posts_root_id_slug ON posts (rootId, slug);
CREATE INDEX IF NOT EXISTS posts_slug ON posts (slug);

CREATE TABLE votes
(
  nickname CITEXT NOT NULL REFERENCES users (nickname),
  tid INT NOT NULL,
  voice INT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS votes_ind ON votes (tid, nickname);

DROP TRIGGER IF EXISTS vote_insertion ON votes;
DROP TRIGGER IF EXISTS vote_updating ON votes;
DROP TRIGGER IF EXISTS add_root_id ON posts;
DROP TRIGGER IF EXISTS thread_insertion ON threads;
DROP TRIGGER IF EXISTS new_thread_author ON threads;

CREATE OR REPLACE FUNCTION insert_vote() RETURNS TRIGGER AS
$vote_insertion$
BEGIN
  UPDATE threads
  SET votes = votes + new.voice
    WHERE id = new.tid;
    RETURN NEW;
END;
$vote_insertion$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_vote() RETURNS TRIGGER AS
$vote_updating$
BEGIN
  UPDATE threads
    SET votes = votes - OLD.voice + NEW.voice
    WHERE id = new.tid;
  RETURN NEW;
END;
$vote_updating$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION init_post() RETURNS TRIGGER AS
$add_root_id$
BEGIN
  UPDATE forums
    SET posts = posts + 1
    WHERE slug = new.forum;
  INSERT INTO forum_users VALUES (new.author, new.forum) ON CONFLICT DO NOTHING;
  RETURN new;
END;
$add_root_id$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION inc_threads() RETURNS TRIGGER AS
$thread_insertion$
BEGIN
  UPDATE forums
    SET threads = threads + 1
    WHERE slug = new.forum;
  RETURN NEW;
END;
$thread_insertion$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION add_forum_user() RETURNS TRIGGER AS
$$
BEGIN
  INSERT INTO forum_users VALUES (new.author, new.forum) ON CONFLICT DO NOTHING;
  RETURN new;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER vote_updating BEFORE UPDATE ON votes FOR EACH ROW EXECUTE PROCEDURE update_vote();
CREATE TRIGGER vote_insertion BEFORE INSERT ON votes FOR EACH ROW EXECUTE PROCEDURE insert_vote();
CREATE TRIGGER add_root_id AFTER INSERT ON posts FOR EACH ROW EXECUTE PROCEDURE init_post();
CREATE TRIGGER thread_insertion AFTER INSERT ON threads FOR EACH ROW EXECUTE PROCEDURE inc_threads();
CREATE TRIGGER new_thread_author AFTER INSERT ON threads FOR EACH ROW EXECUTE PROCEDURE add_forum_user();