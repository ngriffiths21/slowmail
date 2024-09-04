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

// data to pass to the mailbox templates
type mailboxData struct {
    Date string
    Mails []Mail
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
    displayName := req.PostForm.Get("display_name")
    password := sha512.Sum512([]byte(req.PostForm.Get("password")))

    userid, dbErr := newUser(User{Username: username, DisplayName: displayName, Password: password[:]})
    
    if dbErr == ErrNotUnique {
        renderPage(writer, req, signupData{true})
        return
    }
    if dbErr != nil {
        internalError(writer, dbErr)
        return
    }
    if userid == nil {
        internalError(writer, errors.New("No userid was returned by the database."))
        return
    }
    startSession(writer, req, *userid)
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
    if dbErr == ErrNotFound {
        renderPage(writer, req, loginData{true, false, username})
        return
    }
    if dbErr != nil {
        internalError(writer, dbErr)
        return
    }
    if password != [64]byte(user.Password) {
        renderPage(writer, req, loginData{false, true, username})
        return
    }
    startSession(writer, req, user.UserId)
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
    
    // if newSession fails because of duplicate session id, keep trying
    var sessionId string
    randBytes := make([]byte, 8)
    for {
        _, err := rand.Read(randBytes)
        if err != nil {
            internalError(writer, err)
            return
        }
        sessionId = base64.RawStdEncoding.EncodeToString(randBytes)

        err = newSession(Session{sessionId, user, start.Unix(), req.RemoteAddr, expire.Unix()})
        if err == nil {
            break
        } else if err != ErrNotUnique {
            internalError(writer, err)
            return
        }
    }
    http.SetCookie(writer, &http.Cookie{Name: "sessionid", Value: sessionId, Path: "/", Expires: expire})
    http.Redirect(writer, req, "/", http.StatusSeeOther)
}

/* makeAuthedHandler

Returns a handler that checks session authentication and then calls the next handler.
If there is no session cookie, or the session cookie has expired, then it redirects
to the login page.

*/
func makeAuthedHandler(callback func(http.ResponseWriter, *http.Request, int)) func(http.ResponseWriter, *http.Request) {
    return func(writer http.ResponseWriter, req *http.Request) {
        sessionCookie, err := req.Cookie("sessionid")
        if err == http.ErrNoCookie {
            http.Redirect(writer, req, "/login", http.StatusSeeOther)
            return
        }

        var session *Session
        session, err = loadSession(sessionCookie.Value)
        if err != nil && err != ErrNotFound {
            internalError(writer, err)
            return
        }
        if err == ErrNotFound || session.Expiration < time.Now().Unix() {
            http.Redirect(writer, req, "/login", http.StatusSeeOther)
            return
        }
        callback(writer, req, session.UserId)
    }
}

/* getInbox: display inbox */
func getInbox(writer http.ResponseWriter, req *http.Request, user int) {
    mails, err := loadMailbox(user, "inbox")
    if err != nil {
        internalError(writer, err)
        return
    }

    renderPage(writer, req, mailboxData{Date: time.Now().Format("Monday, Jan _2"), Mails: mails})
}

func startServer() error {
    http.HandleFunc("GET /signup", getSignup)
    http.HandleFunc("POST /signup", postSignup)
    http.HandleFunc("GET /login", getLogin)
    http.HandleFunc("POST /login", postLogin)
    http.HandleFunc("GET /mail/folder/inbox", makeAuthedHandler(getInbox))
    http.Handle("GET /{$}", http.RedirectHandler("/mail/folder/inbox", http.StatusSeeOther))

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
