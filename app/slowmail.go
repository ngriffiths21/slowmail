package main

import (
    _ "github.com/mattn/go-sqlite3"
    "database/sql"
    "net/http"
    "html/template"
    "crypto/sha512"
    "flag"
    "fmt"
    "os"
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

var dbPath string

func connectDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		return nil, err
	}

    return db, nil
}

func newUser(username string, display_name string, password []byte) *serverError {
    db, err := connectDb()
    if err != nil {
        return &serverError{"database", err.Error(), ""}
    }
    defer db.Close()

    query := `
        select username from users where username = ?;
    `

    rows, err := db.Query(query, username)
    if err != nil {
        return &serverError{"database", err.Error(), ""}
    }
    /* if rows.Next() is true, the username exists */
    if rows.Next() {
        rows.Close()
        errstr := "This username is already taken."
        return &serverError{"formValidation", errstr, "user"}
    } else { /* if rows.Next() is false, either it failed or is empty */
        rows.Close()
        err = rows.Err()
        if err != nil {
            return &serverError{"database", err.Error(), ""}
        }
    }

    query = `insert into users values (null, ?, ?, ?, ?);`

    _, err = db.Exec(query, username, password, display_name, nil)
    if err != nil {
        return &serverError{"database", err.Error(), ""}
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
