package main

const create_mail = `
	create table if not exists mail (
		mail_id integer primary key,
		user_id integer not null,
		folder varchar(25) check (folder in('inbox', 'archive')),
		read tinyint not null,
		orig_date unsigned int not null,
		date unsigned int not null,
		from_head text not null check (length(from_head) > 0),
		from_name varchar(40),
		from_addr varchar(255) not null check (length(from_addr) > 0),
		to_head text,
		message_id text unique not null,
		in_reply_to text,
		subject text,
		content text,
		multifrom tinyint not null,
		multito tinyint not null
	);
`

const create_users = `
	create table if not exists users (
		user_id integer primary key,
		username varchar(40) unique not null check (length(username) > 0),
		password binary(64) not null,
		display_name varchar(40) not null check (length(display_name) > 0),
		recovery_addr varchar(255)
	);
`

const create_sessions = `
	create table if not exists sessions (
		session_id varchar(11) unique not null,
		user_id integer not null,
		start_date unsigned int not null,
		ip varchar(40) not null,
		expiration unsigned int not null
	);
`

const create_drafts = `
	create table if not exists drafts (
		draft_id integer primary key,
		user_id integer not null,
		recipient varchar(40) not null,
		subject text,
		content text
	);
`

func dbInit() error {
	_, err := db.Exec(create_mail)
	if err != nil {
		return err
	}
	_, err = db.Exec(create_users)
	if err != nil {
		return err
	}
	_, err = db.Exec(create_sessions)
	if err != nil {
		return err
	}
	_, err = db.Exec(create_drafts)
	return err
}
