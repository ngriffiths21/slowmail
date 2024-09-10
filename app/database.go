package main

import (
    "database/sql"
    sqlite "github.com/mattn/go-sqlite3"
    "errors"
)

/* a generic row for loadSingleRow. Note that ToPtrSlice()
method should always have a pointer receiver: if obj is a struct,
obj.ToPtrSlice() receives a *copy* of obj, just like any other struct argument that 
is passed by value.
*/
type DbRowPtr interface {
    ToPtrSlice() []any
}

// mail record
type Mail struct {
    MailId int
    UserId int
    Folder string
    Read bool
    OrigDate int64
    Date int64
    FromHead string
    FromName string
    FromAddr string
    ToHead string
    MessageId string
    InReplyTo string
    Subject string
    Content string
    MultiFrom bool
    MultiTo bool
}

// draft record
type Draft struct {
    UserId int
    Recipient string
    Subject string
    Content string
}

// user record
type User struct {
    UserId int
    Username string
    Password []byte
    DisplayName string
    RecoveryAddr string
}

// session record to retrieve and pass to application
type SessionUser struct {
    SessionId string
    UserId int
    Username string
    DisplayName string
    StartDate int64
    Ip string
    Expiration int64
}

// new session record to pass to db
type Session struct {
    SessionId string
    UserId int
    StartDate int64
    Ip string
    Expiration int64
}

// errors
var (
    ErrNotFound = errors.New("query returned nothing from the database.")
    ErrMultipleRecords = errors.New("query unexpectedly returned more than one record.")
    ErrNotUnique = errors.New("this record already exists.")
)

var db *sql.DB

func (m *Mail) ToPtrSlice() []any {
    return []any{&m.MailId, &m.UserId, &m.Folder, &m.Read, &m.OrigDate, &m.Date, &m.FromHead, &m.FromName,
        &m.FromAddr, &m.ToHead, &m.MessageId, &m.InReplyTo,
        &m.Subject, &m.Content, &m.MultiFrom, &m.MultiTo}
}

func (d *Draft) ToPtrSlice() []any {
    return []any{&d.UserId, &d.Recipient, &d.Subject, &d.Content}
}

func (u *User) ToPtrSlice() []any {
    return []any{&u.UserId, &u.Username, &u.Password, &u.DisplayName, &u.RecoveryAddr}
}

func (s *Session) ToPtrSlice() []any {
    return []any{&s.SessionId, &s.UserId, &s.StartDate, &s.Ip, &s.Expiration}
}

func (s *SessionUser) ToPtrSlice() []any {
    return []any{&s.SessionId, &s.UserId, &s.Username, &s.DisplayName, &s.StartDate, &s.Ip, &s.Expiration}
}

func connectDb(dbPath string) error {
    var err error
	db, err = sql.Open("sqlite3", dbPath)
    return err
}

/* loadSingleRow

Executes a query and expects to find a single row, which it returns.

Parameters:
- query: string SQL query with placeholders for args
- args: args to pass into SQL query
- row: pointer to an initialized struct, which will be filled with query values.

The row struct is updated to hold the results of the query.
If multiple rows are returned, returns ErrMultipleRecords. If nothing is returned,
returns ErrNotFound.

*/
func loadSingleRow(query string, args []any, row DbRowPtr) error {
    rows, err := db.Query(query, args...)
    defer rows.Close()
    if err != nil {
        return err
    }

    // need to call Next() before scanning result
    if !rows.Next() {
        rows.Close()
        err = rows.Err()
        if err == nil {
            err = ErrNotFound
        }
        return err
    }

    // rows.Next found a first row
    err = rows.Scan(row.ToPtrSlice()...)

    // double check there isn't a second result
    if rows.Next() {
        return ErrMultipleRecords
    }

    return err
}

/* newUser: insert a user and return errors

The first return value is user id, generated by the database. The second is an `error`
from the database driver. If the username exists, ErrNotUnique is returned.
If the returned *int is nil, the application failed to get a userId and an error should
have been returned.
*/
func newUser(user User) (*int, error) {
    query := `insert into users values (null, ?, ?, ?, ?);`
    _, err := db.Exec(query, user.Username, user.Password, user.DisplayName, user.RecoveryAddr)
    sqliteErr, _ := err.(sqlite.Error)
    if sqliteErr.ExtendedCode == sqlite.ErrConstraintUnique {
        return nil, ErrNotUnique
    } else if err != nil {
        return nil, err
    }
    rows, err := db.Query("select last_insert_rowid()")
    defer rows.Close()
    if err != nil {
		return nil, err
	}
    if !rows.Next() {
        rows.Close()
        err = rows.Err()
        return nil, err
    }

    var userId int
    err = rows.Scan(&userId)
    if err != nil {
        return nil, err
    }
    return &userId, err
}

/* loadUser

Load a user record. Returns an error if a duplicate is found.
*/
func loadUser(username string) (*User, error) {
    query := `
        select user_id, username, password, display_name, coalesce(recovery_addr, "")
        from users
        where username = ?;
    `

    var user User
    err := loadSingleRow(query, []any{username}, &user)
    if err == ErrNotFound {
        return nil, err
    }
    return &user, err
}

/* newSession: insert a session

The database enforces unique session IDs, ErrNotUnique will be returned if id was
duplicate. This should rarely happen as these should be randomly generated.
When it does happen, the server should simply try again. */
func newSession(session Session) error {
    query := `insert into sessions values (?, ?, ?, ?, ?)`
    _, err := db.Exec(query, session.ToPtrSlice()...)
    sqliteErr, _ := err.(sqlite.Error)
    if sqliteErr.ExtendedCode == sqlite.ErrConstraintUnique {
        err = ErrNotUnique
    }
    return err
}
/* loadSession: select a session

Returns a pointer to the session. If no session is found, returns
a nil pointer.
*/
func loadSession(sessionId string) (*SessionUser, error) {
    query := `
        select session_id, sessions.user_id, username, display_name, start_date, ip, expiration
        from sessions left join users
            on sessions.user_id = users.user_id
        where session_id = ?
    `

    var session SessionUser
    err := loadSingleRow(query, []any{sessionId}, &session)
    if err == ErrNotFound {
        return nil, err
    }
    return &session, err
}

/* newMail

Save a new mail. Returns database driver errors. */
func newMail(mail Mail) error {
    query := `insert into mail values (null, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
    mailFields := mail.ToPtrSlice()[1:] // remove mailId
    _, err := db.Exec(query, mailFields...)
    return err
}

/* loadUserMail: load all mail for a user's mailbox

Params:
- user: Slow Mail user id
- folder: folder (see schema for options)
*/
func loadMailbox(user int, folder string) ([]Mail, error) {
    query := `
        select mail_id, user_id, folder, read, orig_date, date,
            from_head, from_name, from_addr, to_head, message_id, in_reply_to,
            subject, content, multifrom, multito
        from mail
        where user_id = ? and folder = ?;
    `

	rows, err := db.Query(query, user, folder)
    defer rows.Close()
	if err != nil {
		return nil, err
	}

    var mails []Mail
    var mail Mail

    // `Next` must be called even before first row, sets cursor and
    // returns False if none OR if an error occurred
    for rows.Next() {
        // `Scan` copies data from rows to destination
        err = rows.Scan((&mail).ToPtrSlice()...)

        if err != nil {
            return nil, err
        }
        mails = append(mails, mail)
    }

    rows.Close()

    // check if an error happened during `Next()`
    err = rows.Err()
    return mails, err
}

func newDraft(draft Draft) error {
    query := "insert into drafts values (?, ?, ?, ?)"

    _, err := db.Exec(query, draft.ToPtrSlice()...)
    sqliteErr, _ := err.(sqlite.Error)

    if sqliteErr.ExtendedCode == sqlite.ErrConstraintPrimaryKey {
        return ErrNotUnique
    }
    return err
}

func updateDraft(draft Draft) error {
    query := `
        update drafts
        set subject = ?, content = ?
        where user_id = ? and recipient = ?
    `
    _, err := db.Exec(query, draft.Subject, draft.Content, draft.UserId, draft.Recipient)
    return err
}
