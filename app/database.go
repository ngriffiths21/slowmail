package main

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "time"
)

type Mail struct {
    MessageId string
    Date int
    FromName string
    FromAddr string
    MultiFrom bool
    MultiTo bool
    Subject string
    Content string
}

var db *sql.DB

func connectDb(dbPath string) error {
    var err error
	db, err = sql.Open("sqlite3", dbPath)
    return err
}

/* newUser: insert a user and return errors

The first return value is user id. The second is an `error`
originating from the database driver, and the third is a
boolean flag that the user exists. It is set to true if user creation failed
because the username already exists, false otherwise.
*/
func newUser(username string, display_name string, password []byte) (userId int, err error, userExists bool) {
    userId = -1
    query := `insert into users values (null, ?, ?, ?, ?);`
    _, err = db.Exec(query, username, password, display_name, nil)
    if err != nil && err.Error() == "UNIQUE constraint failed: users.username" {
        userExists = true
        err = nil
        return
    }
    rows, err := db.Query("select last_insert_rowid()")
    if err != nil {
		return
	}
    if !rows.Next() {
        err = rows.Err()
        return
    }
    err = rows.Scan(&userId)
    return
}

/* newSession: insert a session

The database enforces unique session IDs, so the second return value is a flag
that the id was duplicate. This should rarely happen as these should be randomly generated.
When it does happen, the server should simply try again. */
func newSession(session string, user int, start_date time.Time, ip string, expiration time.Time) (err error, sessionExists bool) {
    query := `insert into sessions values (?, ?, ?, ?, ?)`
    _, err = db.Exec(query, session, user, start_date, ip, expiration)
    if err != nil && err.Error() == "UNIQUE constraint failed: sessions.session_id" {
        sessionExists = true
        err = nil
        return
    }
    return
}

/* loadUserMail: load all mail for a user

This function should not be used as is. It should be updated
to select a mailbox, not all mail.
*/
func loadUserMail(user int) ([]Mail, error) {
    query := `
        select message_id, date, from_name, from_addr, multifrom,
            multito, subject, content
        from mail
        where user_id = ?
    `

	rows, err := db.Query(query, user)
	if err != nil {
		return nil, err
	}

    var mails []Mail
    mail := Mail{}

    // `Next` must be called even before first row, sets cursor and
    // returns False if none OR if an error occurred
    for rows.Next() {
        // `Scan` copies data from rows to destination
        err = rows.Scan(&mail.MessageId,
                &mail.Date,
                &mail.FromName,
                &mail.FromAddr,
                &mail.MultiFrom,
                &mail.MultiTo,
                &mail.Subject,
                &mail.Content)

        if err != nil {
            return nil, err
        }
        mails = append(mails, mail)
    }

    rows.Close()

    // check if an error happened during `Next()`
    err = rows.Err()
    return mails, err
}
