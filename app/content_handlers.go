package main

import (
    "net/http"
    "log"
    "strings"
    "strconv"
    "time"
    "errors"
    "bytes"
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

    var b bytes.Buffer
    err := temps.ExecuteTemplate(&b, tempName + ".go.tmpl", pdata)
    if (err != nil) {
        internalError(writer, err)
        return
    }
    // buffered so that if an error occurs on template execution, the partial data won't send to client.
    // failing to do this results in an internal error message embedded in a partial HTML page.
    b.WriteTo(writer)
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

/* mailsToPage 

Takes an array of mail, a page number, and a next page number, and returns a truncated array.
*/
func mailsToPage(mails []Mail, page int, next int) []Mail {
    var pageMails []Mail
    if next == 0 {
        pageMails = mails[(page-1)*mailPerPage:]
    } else {
        pageMails = mails[(page-1)*mailPerPage:page*mailPerPage]
    }
    return pageMails
}

func currDate() time.Time {
    currTime := time.Now()
    date := time.Date(currTime.Year(), currTime.Month(), currTime.Day(), 0, 0, 0, 0, time.Local)
    if time.Since(date) < timeOfDelivery {
        // not yet time to deliver today's mail, subtract a day

        // it's simpler to use a duration for timeOfDelivery. technically
        // if we set timeOfDelivery to 11:30pm the mail won't deliver on
        // 23-hour daylight savings days, which is fine.
        date = date.AddDate(0, 0, -1)
    }
    return date
}

/* getInbox: display inbox */
func getInbox(writer http.ResponseWriter, req *http.Request, session SessionUser) {
    inboxDate := currDate()

    mails, err := loadInbox(session.UserId, inboxDate.Unix())
    if err != nil {
        internalError(writer, err)
        return
    }

    page, next := parsePages(req, len(mails))

    pageMails := mailsToPage(mails, page, next)

    // truncate the content of mail and construct previews
    var previews []mailPreview

    for _, m := range pageMails {
        preview := mailPreview{MailId: m.MailId, FromName: m.FromName, Subject: m.Subject, Preview: trunc(m.Content, 60)}
        previews = append(previews, preview)
    }

    renderPage(writer, req, mailboxData{Username: session.Username, Date: inboxDate.Format("Monday, Jan 2"), Mails: previews,
        PagePrev: page - 1, PageNext: next})
}

/* getConv 

Parses the request path to get a mail ID. Loads the sender associated with that mail.
Then renders the conversation page, including all mail with that sender, as well as
any saved draft to that sender.
*/
func getConv(writer http.ResponseWriter, req *http.Request, session SessionUser) {
    mailId := req.PathValue("mailId")
    if mailId == "" {
        internalError(writer, errors.New("Could not parse mail id from conversation path"))
        return
    }
    sender, err := loadSenderAddr(mailId)
    if err != nil {
        internalError(writer, err)
        return
    }

    convDate := currDate().Unix()

    mails, err := loadConv(session.UserId, sender.SenderAddr, convDate)
    if err != nil {
        internalError(writer, err)
        return
    }
    draft, err := loadDraft(session.UserId, sender.SenderAddr)
    if err != nil && err != ErrNotFound {
        internalError(writer, err)
        return
    }

    page, next := parsePages(req, len(mails))
    var draftDisplay *mailDisplay
    if page == 1 && draft != nil {
        // only show the draft on first page
        draftDisplay = &mailDisplay{Subject: draft.Subject, Content: draft.Content}
    }
    pageMails := mailsToPage(mails, page, next)

    var displayMails []mailDisplay

    for _, m := range pageMails {
        display := mailDisplay{Date: time.Unix(m.Date, 0).Format("Monday, Jan 2, 2006"), Subject: m.Subject, Content: m.Content}
        displayMails = append(displayMails, display)
    }

    renderPage(writer, req, convData{Username: session.Username, MailId: mailId, SenderName: sender.SenderName, SenderAddr: sender.SenderAddr,
        Draft: draftDisplay, Mails: displayMails, PagePrev: page - 1, PageNext: next})
}

func getCompose(writer http.ResponseWriter, req *http.Request, session SessionUser) {
    renderPage(writer, req, composeData{Username: session.Username})
}
