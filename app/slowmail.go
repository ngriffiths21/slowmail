package main

import (
    "net/http"
    "html/template"
    "crypto/sha512"
    "flag"
    "log"
    "os"
)

type signupData struct {
    UserExists bool
}

func internalError(writer http.ResponseWriter, err error) {
    log.Println(err.Error())
    http.Error(writer, err.Error(), http.StatusInternalServerError)
}

func renderSignup(writer http.ResponseWriter, sdata signupData) {
    t, err := template.New("new.go.tmpl").ParseFiles("templates/pages/account/new.go.tmpl",
        "templates/css/styles.go.tmpl")

    if (err != nil) {
        internalError(writer, err)
    }
    err = t.Execute(writer, sdata)
    if (err != nil) {
        internalError(writer, err)
    }
}

func getSignup(writer http.ResponseWriter, req *http.Request) {
    renderSignup(writer, signupData{false})
}

func newAccount(writer http.ResponseWriter, req *http.Request) {
    err := req.ParseForm()
    if (err != nil) {
        internalError(writer, err)
    }

    username := req.PostForm.Get("username")
    display_name := req.PostForm.Get("display_name")
    password := sha512.Sum512([]byte(req.PostForm.Get("password")))

    dbErr, userExists := newUser(username, display_name, password[:])
    if dbErr != nil {
        internalError(writer, dbErr)
    } else if userExists {
        renderSignup(writer, signupData{true})
    } else {
        http.Redirect(writer, req, "/", http.StatusSeeOther)
    }
}

func startServer() error {
    http.HandleFunc("GET /account/new", getSignup)
    http.HandleFunc("POST /account/new", newAccount)

    err := http.ListenAndServe(":8080", nil)
    return err
}

func main() {
    var dbPath string
    flag.StringVar(&dbPath, "db", "", "Path to the database")
    flag.Parse()
    if dbPath == "" {
        log.Println("Error: no path to database provided.")
        flag.Usage()
        os.Exit(1)
    }

    err := connectDb(dbPath)
    if (err != nil) {
        log.Panic(err)
    }
    defer db.Close()

    err = startServer()
    if (err != nil) {
        log.Panic(err)
    }
}
