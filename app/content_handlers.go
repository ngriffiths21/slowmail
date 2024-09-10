package main

import (
    "net/http"
    "log"
    "strings"
    "strconv"
    "time"
)

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

/* parsePages

Parses the page query parameter, and returns two ints: the current page, and the next page.
If an invalid current page is passed as a query parameter, this function returns page 1.
If the current page is also the last, `next`` will be set to 0.
*/
func parsePages(req *http.Request, mailcount int) (int, int) {
    // FormValue ignores parse errors which is desired behavior in this case
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
        preview := mailPreview{MailId: m.MailId, FromName: m.FromName, Subject: m.Subject, Preview: trunc(m.Content, 60)}
        previews = append(previews, preview)
    }

    renderPage(writer, req, mailboxData{Username: session.Username, Date: time.Now().Format("Monday, Jan _2"), Mails: previews,
        PagePrev: page - 1, PageNext: next})
}

func getCompose(writer http.ResponseWriter, req *http.Request, session SessionUser) {
    renderPage(writer, req, composeData{Username: session.Username})
}
