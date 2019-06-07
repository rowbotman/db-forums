--CREATE DATABASE park_forum WITH OWNER = postgres;

CREATE EXTENSION IF NOT EXISTS CITEXT;

CREATE EXTENSION IF NOT EXISTS CITEXT;
TRUNCATE TABLE profile CASCADE;
TRUNCATE TABLE forum CASCADE;
TRUNCATE TABLE thread CASCADE;
TRUNCATE TABLE vote CASCADE;
TRUNCATE TABLE post CASCADE;

DROP TABLE IF EXISTS profile CASCADE;
DROP TABLE IF EXISTS forum CASCADE;
DROP TABLE IF EXISTS thread CASCADE;
DROP TABLE IF EXISTS post CASCADE;
DROP TABLE IF EXISTS vote CASCADE;
-- DROP TABLE IF EXISTS  CASCADE;
--CREATE ROLE park_forum;

CREATE TABLE IF NOT EXISTS profile
(
  uid       SERIAL PRIMARY KEY,
  nickname  CITEXT  UNIQUE NOT NULL CHECK (nickname <> ''),
  full_name VARCHAR(128)        NOT NULL CHECK (nickname <> ''),
  about     VARCHAR(512),
  email     VARCHAR(256) UNIQUE NOT NULL CHECK (email <> '')
);

CREATE TABLE IF NOT EXISTS forum
(
  uid       SERIAL PRIMARY KEY,
  title     VARCHAR(128)        NOT NULL CHECK ( title <> '' ),
  author_id INT                 NOT NULL,
  slug      VARCHAR(256) UNIQUE NOT NULL,

  FOREIGN   KEY (author_id) REFERENCES profile (uid)
);

CREATE TABLE IF NOT EXISTS thread
(
  uid      SERIAL PRIMARY KEY,
  user_id  INT NOT NULL,
  forum_id INT NOT NULL,
  title    VARCHAR(128)                NOT NULL CHECK ( title <> '' ),
  slug     VARCHAR(128) UNIQUE         NOT NULL,
  message  VARCHAR(2048)               NOT NULL,
  created  TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  votes    INT                         NOT NULL DEFAULT 0,

  FOREIGN  KEY (user_id)  REFERENCES profile (uid),
  FOREIGN  KEY (forum_id) REFERENCES forum   (uid)
);

CREATE TABLE IF NOT EXISTS post
(
  uid       SERIAL PRIMARY KEY,
  parent_id INT                    DEFAULT 0,
  path      INTEGER[]    NOT NULL,
  forum_id  INT          NOT NULL,
  user_id   INT          NOT NULL,
  thread_id INT          NOT NULL,
  message   VARCHAR(512) NOT NULL,
  is_edited BOOLEAN                DEFAULT FALSE,
  created   TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT current_timestamp,

  FOREIGN   KEY (forum_id)  REFERENCES forum   (uid),
  FOREIGN   KEY (user_id)   REFERENCES profile (uid),
  FOREIGN   KEY (parent_id) REFERENCES post    (uid),
  FOREIGN   KEY (thread_id) REFERENCES thread  (uid)
);

CREATE TABLE IF NOT EXISTS vote
(
  user_id   INT           NOT NULL,
  thread_id INT           NOT NULL,
  value     SMALLINT      NOT NULL   DEFAULT 0,

  FOREIGN KEY (user_id)   REFERENCES profile (uid) ON DELETE CASCADE,
  FOREIGN KEY (thread_id) REFERENCES post    (uid) ON DELETE CASCADE
);

-- GRANT ALL PRIVILEGES ON DATABASE park_forum TO park_forum;--why we granted privileges to park_forum if
--                                                           --db owner is postgres?
-- GRANT USAGE ON SCHEMA public TO park_forum;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO park_forum;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO park_forum;
