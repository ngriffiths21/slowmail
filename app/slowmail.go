package main

import (
    "net/http"
    "html/template"
    "crypto/sha512"
    "flag"
    "log"
    "os"
    "strings"
)

// data to pass to the signup page template
type signupData struct {
    UserExists bool
}

// parsed templates, will be initialized on app init
var temps *template.Template

/* internalError

Logs error and sends 500. Use for any error that is an application bug
or a system error, not a user error.
*/
func internalError(writer http.ResponseWriter, err error) {
    log.Println(err.Error())
    http.Error(writer, "The server encountered an error and couldn't fulfill this request. Sorry about that.",
        http.StatusInternalServerError)
}

/* renderPage

Checks the request path and chooses the template that matches the last part of the request path.
`pdata` is page data, and the proper type depends on the template being rendered.
*/
func renderPage(writer http.ResponseWriter, req *http.Request, pdata any) {
    url := req.URL.Path
    // if last character is '/', remove it
    if url[len(url) - 1] == '/' {
        url = url[:len(url) - 1]
    }
    // extract last part of path
    tempName := url[strings.LastIndex(url, "/") + 1:]
    err := temps.ExecuteTemplate(writer, tempName + ".go.tmpl", pdata)
    if (err != nil) {
        internalError(writer, err)
    }
}

func getSignup(writer http.ResponseWriter, req *http.Request) {
    renderPage(writer, req, signupData{false})
}

/* postSignup

Parses a signup form and inserts a new user to the database.
This route handles duplicate username as a user error, and all
others as internal errors (form validation is the client's
responsibility).
*/
func postSignup(writer http.ResponseWriter, req *http.Request) {
    err := req.ParseForm()
    if (err != nil) {
        internalError(writer, err)
        return
    }

    username := req.PostForm.Get("username")
    display_name := req.PostForm.Get("display_name")
    password := sha512.Sum512([]byte(req.PostForm.Get("password")))

    dbErr, userExists := newUser(username, display_name, password[:])
    if dbErr != nil {
        internalError(writer, dbErr)
        return
    } else if userExists {
        renderPage(writer, req, signupData{true})
        return
    }
    http.Redirect(writer, req, "/", http.StatusSeeOther)
}

func startServer() error {
    http.HandleFunc("GET /account/new", getSignup)
    http.HandleFunc("POST /account/new", postSignup)

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

    var err error
    temps, err = template.ParseFiles("templates/pages/account/new.go.tmpl",
        "templates/css/styles.go.tmpl")
    if (err != nil) {
        log.Panic(err)
    }

    err = connectDb(dbPath)
    if (err != nil) {
        log.Panic(err)
    }
    defer db.Close()

    err = startServer()
    if (err != nil) {
        log.Panic(err)
    }
}
