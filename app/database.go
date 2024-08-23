package main

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
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

var dbPath string

func connectDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		return nil, err
	}

    return db, nil
}

func newUser(username string, display_name string, password []byte) *dbError {
    db, err := connectDb()
    if err != nil {
        return &dbError{"database", err.Error()}
    }
    defer db.Close()

    query := `insert into users values (null, ?, ?, ?, ?);`

    _, err = db.Exec(query, username, password, display_name, nil)
    if err != nil {
        if err.Error() == "UNIQUE constraint failed: users.username" {
            return &dbError{"userExists", err.Error()}
        }
        return &dbError{"database", err.Error()}
    }

    return nil
}

func loadUserMail(user int) ([]Mail, error) {
    db, err := connectDb()
    if err != nil {
        return nil, err
    }
	defer db.Close()

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
    if err != nil {
        return nil, err
    }

    return mails, nil
}