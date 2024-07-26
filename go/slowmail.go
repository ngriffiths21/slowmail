package main

import (
    "fmt"
    _ "github.com/mattn/go-sqlite3"
    "database/sql"
    "log"
)

type Mail struct {
    Id int
    Date int
    Content string
}

func loadFirstMail() Mail {
	db, err := sql.Open("sqlite3", "./sqlite/mail.db")

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	rows, err := db.Query("select * from mail")
	if err != nil {
		log.Panic(err)
	}

    defer rows.Close()

    rows.Next() // prep the first row

    var id int
    var date int
    var content string
    
    err = rows.Scan(&id, &date, &content) // save data
    if err != nil {
		log.Panic(err)
	}

    return Mail{Id: id, Date: date, Content: content}
}

func main() {
    fmt.Println(loadFirstMail())
}
