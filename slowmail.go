package main

import (
    _ "github.com/mattn/go-sqlite3"
    "database/sql"
    "net/http"
    "html/template"
    "log"
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

func getSignup(writer http.ResponseWriter, req *http.Request) {
    t, err := template.New("new.go.tmpl").ParseFiles("templates/pages/account/new.go.tmpl",
        "templates/css/styles.go.tmpl")
    if (err != nil) {
        panic(err)
    }
    err = t.Execute(writer, nil)
    if (err != nil) {
        panic(err)
    }
}

func startServer() error {
    http.HandleFunc("/account/new", getSignup)

    err := http.ListenAndServe(":8080", nil)
    return err
}

func main() {
    log.Fatal(startServer())
}
