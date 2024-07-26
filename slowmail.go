package main

import (
    "fmt"
    _ "github.com/mattn/go-sqlite3"
    "database/sql"
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

func connectDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "../sqlite/mail.db")

	if err != nil {
		return nil, err
	}

    return db, nil
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
        // this scan fails the first time with seg fault
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

func main() {
    fmt.Println("Doesn't do anything yet. Try running the tests.")
}
