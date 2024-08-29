package main

import (
    "net/http"
    "html/template"
    "crypto/sha512"
    "crypto/rand"
    "encoding/base64"
    "flag"
    "log"
    "os"
    "strings"
    "errors"
    "time"
)

// data to pass to the signup page template
type signupData struct {
    UserExists bool
}

// data to pass to the login page template
type loginData struct {
    UserWrong bool
    PassWrong bool
    Username string
}

// user record
type user struct {
    userId int
    username string
    password []byte
    displayName string
    recoveryAddr string
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

func getLogin(writer http.ResponseWriter, req *http.Request) {
    renderPage(writer, req, loginData{false, false, ""})
}

/* postSignup

Parses a signup form and inserts a new user to the database.
This route handles duplicate username as a user error, and all
others as internal errors (form validation is the client's
responsibility).
*/
func postSignup(writer http.ResponseWriter, req *http.Request) {
    err := req.ParseForm()
    if err != nil {
        internalError(writer, err)
        return
    }

    username := req.PostForm.Get("username")
    display_name := req.PostForm.Get("display_name")
    password := sha512.Sum512([]byte(req.PostForm.Get("password")))

    userid, dbErr, userExists := newUser(username, display_name, password[:])
    if dbErr != nil {
        internalError(writer, dbErr)
        return
    } else if userExists {
        renderPage(writer, req, signupData{true})
        return
    } else if userid == -1 {
        internalError(writer, errors.New("No userid was returned by the database."))
        return
    }
    startSession(writer, req, userid)
}

/* postLogin

This route handles missing username and missing passwords as user errors.
All others are internal errors.
*/
func postLogin(writer http.ResponseWriter, req *http.Request) {
    err := req.ParseForm()
    if err != nil {
        internalError(writer, err)
        return
    }

    username := req.PostForm.Get("username")
    password := sha512.Sum512([]byte(req.PostForm.Get("password")))

    user, dbErr := loadUser(username)
    if dbErr != nil {
        internalError(writer, dbErr)
        return
    } else if user == nil {
        renderPage(writer, req, loginData{true, false, username})
        return
    } else if password != [64]byte(user.password) {
        renderPage(writer, req, loginData{false, true, username})
    }
    startSession(writer, req, user.userId)
}

/* startSession

Create a new session, save it to the database, set auth cookie, and redirect.
*/
func startSession(writer http.ResponseWriter, req *http.Request, user int) {
    start := time.Now()
    d, err := time.ParseDuration("24h")
    if err != nil {
        internalError(writer, err)
        return
    }
    expire := start.Add(d)
    
    sessionNotStarted := true
    // if newSession fails because of duplicate session id, keep trying
    var sessionId string
    randBytes := make([]byte, 8)
    for sessionNotStarted {
        _, err := rand.Read(randBytes)
        if err != nil {
            internalError(writer, err)
            return
        }
        sessionId = base64.RawStdEncoding.EncodeToString(randBytes)

        err, sessionNotStarted = newSession(sessionId, user, start, req.RemoteAddr, expire)
        if err != nil {
            internalError(writer, err)
            return
        }
    }
    http.SetCookie(writer, &http.Cookie{Name: "sessionid", Value: sessionId, Path: "/", Expires: expire})
    http.Redirect(writer, req, "/", http.StatusSeeOther)
}

func startServer() error {
    http.HandleFunc("GET /signup", getSignup)
    http.HandleFunc("POST /signup", postSignup)
    http.HandleFunc("GET /login", getLogin)
    http.HandleFunc("POST /login", postLogin)

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
    temps, err = template.ParseGlob("templates/*.go.tmpl")
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
