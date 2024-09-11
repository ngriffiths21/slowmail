package main

/* Data types that model application state */

// data to pass to the signup page template
type signupData struct {
	UserExists bool
}

// data to pass to the login page template
type loginData struct {
	UserWrong bool
	PassWrong bool
	Username  string
}

// data to pass to the mailbox templates
type mailboxData struct {
	Username string
	Date     string
	Mails    []mailPreview
	PagePrev int
	PageNext int
}

type mailPreview struct {
	MailId   int
	FromName string
	Subject  string
	Preview  string
}

type mailDisplay struct {
	Date    string
	Subject string
	Content string
}

// data for the conversation view page
type convData struct {
	Username   string
	MailId     string
	SenderName string
	SenderAddr string
	Draft      *mailDisplay
	Mails      []mailDisplay
	PagePrev   int
	PageNext   int
}

//data for compose page
type composeData struct {
	Username string
}
