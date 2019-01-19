-- DROP TABLE IF EXISTS ForumUser, Vote, Thread, Post, Forum, Users CASCADE;
CREATE EXTENSION IF NOT EXISTS citext;
SET LOCAL synchronous_commit TO OFF;

--Users
CREATE TABLE IF NOT EXISTS Users(
	about CITEXT,
	email CITEXT NOT NULL UNIQUE,
	fullname CITEXT NOT NULL,
	nickname CITEXT COLLATE "ucs_basic" NOT NULL UNIQUE
);
CREATE INDEX IF NOT EXISTS usrNickname ON Users(nickname);

--Forums
CREATE TABLE IF NOT EXISTS Forum(
	posts BIGINT DEFAULT 0,
	slug CITEXT NOT NULL UNIQUE,
	threads INTEGER DEFAULT 0,
	title CITEXT NOT NULL,
	author CITEXT COLLATE "ucs_basic" NOT NULL REFERENCES Users(nickname)
);

CREATE INDEX IF NOT EXISTS forumSlug ON Forum(slug);

--Threads
CREATE TABLE IF NOT EXISTS Thread(
	id SERIAL PRIMARY KEY,
	author CITEXT COLLATE "ucs_basic" NOT NULL REFERENCES Users(nickname),
	created TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
	forum CITEXT NOT NULL REFERENCES Forum(slug),
	message CITEXT NOT NULL,
	slug CITEXT UNIQUE,
	title CITEXT NOT NULL,
	votes INTEGER DEFAULT 0
);

CREATE OR REPLACE FUNCTION create_thread() RETURNS TRIGGER AS '
  BEGIN
    UPDATE Forum SET threads=threads+1 WHERE slug=NEW.forum;
    RETURN NEW;
  END;
'
LANGUAGE plpgsql;


CREATE TRIGGER create_thread
BEFORE INSERT ON Thread FOR EACH ROW
EXECUTE PROCEDURE create_thread();


CREATE INDEX IF NOT EXISTS thrSlug ON Thread(slug);
CREATE INDEX IF NOT EXISTS thrId ON Thread(id);
CREATE INDEX IF NOT EXISTS thrForm_athr ON Thread(forum,author);
CREATE INDEX IF NOT EXISTS thrForum_cr ON Thread(forum,created);

--Posts
CREATE TABLE IF NOT EXISTS Post(
	author CITEXT COLLATE "ucs_basic" NOT NULL REFERENCES Users (nickname),
	created TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
	forum CITEXT REFERENCES Forum(slug),
	id BIGSERIAL PRIMARY KEY,
	isedited BOOLEAN DEFAULT FALSE,
	message text NOT NULL,
	parent BIGINT DEFAULT 0,
	thread INTEGER NOT NULL REFERENCES Thread(id),
	id_array BIGINT ARRAY DEFAULT '{}'
);


CREATE OR REPLACE FUNCTION msg_check() RETURNS TRIGGER AS '
  BEGIN
    NEW.isedited:=false;
    RETURN NEW;
  END;
'
LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION create_post() RETURNS TRIGGER AS '
  BEGIN
    IF NEW.parent<>0 AND NOT EXISTS (SELECT id FROM Post WHERE id=NEW.parent AND thread=NEW.thread) THEN
      RAISE ''Parent post exc'';
    END IF;
    NEW.id_array=array_append((SELECT id_array FROM Post WHERE id=NEW.parent), NEW.id);
    UPDATE Forum SET posts=posts+1 WHERE slug=NEW.forum;
    RETURN NEW;
  END;
'
LANGUAGE plpgsql;

CREATE TRIGGER change_message
BEFORE UPDATE ON Post FOR EACH ROW WHEN (new.message=old.message)
EXECUTE PROCEDURE msg_check();

CREATE TRIGGER create_post
BEFORE INSERT ON Post FOR EACH ROW
EXECUTE PROCEDURE create_post();

CREATE INDEX IF NOT EXISTS postId_cr ON Post (id, created);
CREATE INDEX IF NOT EXISTS postThread_id_cr ON Post (thread, id, created);
CREATE INDEX IF NOT EXISTS postParent_thread ON Post (parent, thread);
CREATE INDEX IF NOT EXISTS postId_array ON Post (thread, (id_array[0]), id_array);

--Votes
CREATE TABLE IF NOT EXISTS Vote (
	nickname CITEXT COLLATE "ucs_basic" NOT NULL REFERENCES Users(nickname),
	voice INTEGER NOT NULL,
	thread INTEGER NOT NULL REFERENCES Thread(id),
	UNIQUE (nickname, thread)
);


CREATE OR REPLACE FUNCTION create_vote() RETURNS TRIGGER AS'
  BEGIN
    UPDATE Thread SET votes=votes+NEW.voice WHERE id=NEW.thread;
    RETURN NEW;
  END;
'
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_vote() RETURNS TRIGGER AS'
  BEGIN
    IF (OLD.voice<>NEW.voice) THEN
      IF (NEW.voice=-1) THEN
        UPDATE Thread SET votes=votes-2 WHERE id=NEW.thread;
      ELSE
        UPDATE Thread SET votes=votes+2 WHERE id=NEW.thread;
      END IF;
    END IF;
    RETURN OLD;
  END;
'
LANGUAGE plpgsql;

CREATE TRIGGER create_vote
AFTER INSERT ON Vote FOR EACH ROW
EXECUTE PROCEDURE create_vote();

CREATE TRIGGER update_vote
AFTER UPDATE ON Vote FOR EACH ROW
EXECUTE PROCEDURE update_vote();


--ForumUsers
CREATE TABLE IF NOT EXISTS ForumUser (
  forum CITEXT REFERENCES Forum(slug),
  author CITEXT REFERENCES Users(nickname),
  UNIQUE (forum,author)
);

CREATE INDEX IF NOT EXISTS frmUsrs ON ForumUser (forum, author);



