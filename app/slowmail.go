package main

import (
    "net/http"
    "html/template"
    "crypto/sha512"
    "flag"
    "fmt"
    "os"
)

/* serverError

Fields:
- errorType: "database" for a database IO error,
    or "formValidation" for a form input error
- message: description of the error
- invalidFormField: if formValidation error, which field
    was invalid
*/
type serverError struct {
    errorType string
    message string
    invalidFormField string
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

    serveErr := newUser(username, display_name, password[:])
    if serveErr != nil {
        if serveErr.errorType == "database" {
            http.Error(writer, serveErr.message, http.StatusInternalServerError)
        } else if serveErr.errorType == "formValidation" {
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
