package main

import (
    "net/http"
    "time"
    "crypto/sha512"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "strings"
)

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

func postComposeSave(writer http.ResponseWriter, req *http.Request, session SessionUser) {
    err := req.ParseForm()
    if err != nil {
        internalError(writer, err)
        return
    }

    draft := Draft{Recipient: req.PostForm.Get("to"), Subject: req.PostForm.Get("subject"),
        Content: req.PostForm.Get("content")}
    err = newDraft(draft)
    if err == ErrNotUnique {
        // this err is not reportable, it is app state
        // so it is safe to just reassign err
        err = updateDraft(draft)
    }

    if err != nil {
        internalError(writer, err)
        return
    }

    http.Redirect(writer, req, "/mail/folder/inbox", http.StatusSeeOther)
}

func postComposeSend(writer http.ResponseWriter, req *http.Request, session SessionUser) {
    err := req.ParseForm()
    if err != nil {
        internalError(writer, err)
        return
    }

    recipient, recipientHost, hasAt := strings.Cut(req.PostForm.Get("to"), "@")
    if !hasAt {
        internalError(writer, errors.New("Error: malformed recipient email address"))
        return
    }
    subject := req.PostForm.Get("subject")
    content := req.PostForm.Get("content")

    // first check if recipient exists
    var user *User
    if recipientHost == host {
        user, err = loadUser(recipient)
    }
    var recipientId int
    if recipientHost != host || err == ErrNotFound {
		// recipient does not exist. Change mail to bounce back to sender.
        recipientId = session.UserId
        subject = "Not sent: " + subject
        content = messageNotSent + "Recipient: " + recipient + "\n\n" + content
    } else if err != nil {
        internalError(writer, err)
        return
    } else {
		// recipient found; set recipient ID
        recipientId = user.UserId
    }

    currTime := time.Now()
    currDate := time.Date(currTime.Year(), currTime.Month(), currTime.Day(), 0, 0, 0, 0, time.Local)
    name := session.DisplayName
    addr := session.Username + "@" + host

    mail := Mail{UserId: recipientId,
        Folder: "inbox",
        Read: false,
        OrigDate: currTime.Unix(),
        Date: currDate.Unix(),
        FromHead: name + " <" + addr + ">",
        FromName: name,
        FromAddr: addr,
        ToHead: "",
        MessageId: "",
        InReplyTo: "",
        Subject: subject,
        Content: content,
        MultiFrom: false,
        MultiTo: false}
    
    err = newMail(mail)
    if err != nil {
        internalError(writer, err)
        return
    }

    http.Redirect(writer, req, "/mail/folder/inbox", http.StatusSeeOther)
}
