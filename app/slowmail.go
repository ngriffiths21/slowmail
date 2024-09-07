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
    "strconv"
)

// parsed templates, will be initialized on app init
var temps *template.Template

// page length for mailboxes
var mailPerPage = 12

/* trunc

Safely truncate strings. It does not destroy unicode characters,
but display width may vary. */
func trunc(s string, n int) string {
    runes := []rune(s)
    if len(runes) <= n {
        return s
    }
    return string(runes[0:n])
}

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
func makeAuthedHandler(callback func(http.ResponseWriter, *http.Request, SessionUser)) func(http.ResponseWriter, *http.Request) {
    return func(writer http.ResponseWriter, req *http.Request) {
        sessionCookie, err := req.Cookie("sessionid")
        if err == http.ErrNoCookie {
            http.Redirect(writer, req, "/login", http.StatusSeeOther)
            return
        }

        var session *SessionUser
        session, err = loadSession(sessionCookie.Value)
        if err != nil && err != ErrNotFound {
            internalError(writer, err)
            return
        }
        if err == ErrNotFound || session.Expiration < time.Now().Unix() {
            http.Redirect(writer, req, "/login", http.StatusSeeOther)
            return
        }

        callback(writer, req, *session)
    }
}

func parsePages(req *http.Request, mailcount int) (int, int) {
    page, err := strconv.Atoi(req.FormValue("page"))
    
    // if no page or invalid page, reset to 1
    if err != nil || page < 1 || (page-1)*mailPerPage >= mailcount {
        page = 1
    }
    if page*mailPerPage > mailcount {
        return page, 0
    }
    return page, page + 1
}

/* getInbox: display inbox */
func getInbox(writer http.ResponseWriter, req *http.Request, session SessionUser) {
    mails, err := loadMailbox(session.UserId, "inbox")
    if err != nil {
        internalError(writer, err)
        return
    }

    page, next := parsePages(req, len(mails))

    // truncate the list of mails
    var pageMails []Mail
    if next == 0 {
        pageMails = mails[(page-1)*mailPerPage:]
    } else {
        pageMails = mails[(page-1)*mailPerPage:page*mailPerPage]
    }

    // truncate the content of mail and construct previews
    var previews []mailPreview

    for _, m := range pageMails {
        preview := mailPreview{MailId: m.MailId, FromName: m.FromName, Subject: m.Subject, Preview: trunc(m.Content, 40)}
        previews = append(previews, preview)
    }

    renderPage(writer, req, mailboxData{Username: session.Username, Date: time.Now().Format("Monday, Jan _2"), Mails: previews,
        PagePrev: page - 1, PageNext: next})
}

func getCompose(writer http.ResponseWriter, req *http.Request, session SessionUser) {
    renderPage(writer, req, composeData{Username: session.Username})
}

func startServer() error {
    http.HandleFunc("GET /signup", getSignup)
    http.HandleFunc("POST /signup", postSignup)
    http.HandleFunc("GET /login", getLogin)
    http.HandleFunc("POST /login", postLogin)
    http.HandleFunc("GET /mail/folder/inbox", makeAuthedHandler(getInbox))
    http.HandleFunc("GET /mail/compose", makeAuthedHandler(getCompose))
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
