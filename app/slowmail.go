package main

import (
    "net/http"
    "html/template"
    "crypto/sha512"
    "flag"
    "fmt"
    "os"
)

/* dbError

Fields:
- errorType:
    - "database" for a database IO error
    - "userExists" for a duplicate username
- message: description of the error
*/
type dbError struct {
    errorType string
    message string
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

func newAccount(writer http.ResponseWriter, req *http.Request) {
    err := req.ParseForm()
    if (err != nil) {
        panic(err)
    }

    username := req.PostForm.Get("username")
    display_name := req.PostForm.Get("display_name")
    password := sha512.Sum512([]byte(req.PostForm.Get("password")))

    dbErr := newUser(username, display_name, password[:])
    if dbErr != nil {
        if dbErr.errorType == "database" {
            http.Error(writer, dbErr.message, http.StatusInternalServerError)
        } else if dbErr.errorType == "userExists" {
            http.Redirect(writer, req, "/account/new?error=userexists",
                http.StatusSeeOther)
        }
    }

    http.Redirect(writer, req, "/", http.StatusSeeOther)
}

func startServer() error {
    http.HandleFunc("GET /account/new", getSignup)
    http.HandleFunc("POST /account/new", newAccount)

    err := http.ListenAndServe(":8080", nil)
    return err
}

func main() {
    flag.StringVar(&dbPath, "db", "", "Path to the database")
    flag.Parse()
    if dbPath == "" {
        fmt.Println("Error: no path to database provided. Usage:")
        flag.PrintDefaults()
        os.Exit(1)
    }

    err := startServer()
    if (err != nil) {
        panic(err)
    }
}
