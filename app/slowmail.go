package main

import (
    "net/http"
    "html/template"
    "flag"
    "log"
    "os"
)

// notice that message couldn't be sent

const messageNotSent = "**Sorry, the below message could not be sent:**\n\n"

// parsed templates, will be initialized on app init
var temps *template.Template

// page length for mailboxes
var mailPerPage = 12

// host name for email addresses
var host string

func startServer() error {
    http.HandleFunc("GET /signup/{$}", getSignup)
    http.HandleFunc("POST /signup/{$}", postSignup)
    http.HandleFunc("GET /login/{$}", getLogin)
    http.HandleFunc("POST /login/{$}", postLogin)
    http.HandleFunc("GET /logout/{$}", logout)
    http.HandleFunc("GET /mail/folder/inbox/{$}", makeAuthedHandler(getInbox))
    http.HandleFunc("GET /mail/compose/{$}", makeAuthedHandler(getCompose))
    http.HandleFunc("POST /mail/compose/send/{$}", makeAuthedHandler(postComposeSend))
    http.HandleFunc("POST /mail/compose/{$}", makeAuthedHandler(postComposeSave))
    http.HandleFunc("GET /mail/conv/{mailId}/read/{$}", makeAuthedHandler(getConv))
    http.HandleFunc("POST /mail/conv/{mailId}/send/{$}", makeAuthedHandler(postComposeSend))
    http.HandleFunc("POST /mail/conv/{mailId}/save/{$}", makeAuthedHandler(postComposeSave))
    http.Handle("GET /{$}", http.RedirectHandler("/mail/folder/inbox", http.StatusSeeOther))

    err := http.ListenAndServe(":8080", nil)
    return err
}

func appInit() {
    var dbPath string
    flag.StringVar(&dbPath, "db", "", "Path to the database (required)")
    flag.StringVar(&host, "host", "", "Host name for email addresses (required)")
    flag.Parse()
    if dbPath == "" || host == "" {
        log.Println("Error: please provide all required flags.")
        flag.Usage()
        os.Exit(1)
    }

    var err error
    temps, err = template.ParseGlob("templates/*.go.tmpl")
    if (err != nil) {
        log.Panic(err)
    }

    err = connectDb(dbPath)
    if (err != nil) {
        log.Panic(err)
    }
    // Connect sequentially to avoid write access conflicts
    db.SetMaxOpenConns(1) // it is slow mail after all
}

func main() {
    appInit()
    defer db.Close()
    err := startServer()
    if (err != nil) {
        log.Panic(err)
    }
}
