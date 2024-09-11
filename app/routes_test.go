package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// The tests make the following assumptions about the test database (see error messages):
func checkUser(t *testing.T) {
	user, err := loadUser("test")
	if err != nil || user.UserId != 1 {
		t.Error("Testing requires access to a user 'test', with password 'test' and user_id 1. See test file.")
	}
}

func checkSession(t *testing.T) {
	session, err := loadSession("1")
	if err != nil || session.UserId != 1 {
		t.Error("Testing requires access to a session with sessionid 1 for user 'test'. See test file.")
	}
}

func checkMail(t *testing.T) {
	mail, err := loadMailArray("select * from mail where mail_id = 1 and user_id = 1", []any{})
	if err != nil || len(mail) == 0 {
		t.Error("Testing requires access to a mail with mail_id 1 and user_id 1. See test file.")
	}
}

/* TESTS */

func TestGetSignup(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/signup/", nil)
	getSignup(rw, req)
	if rw.Code != 200 {
		t.Errorf("Expected status 200; got %d", rw.Code)
	}
}

func TestPostSignup(t *testing.T) {
	query := "delete from users where username = 'test';"
	_, err := db.Exec(query)
	if err != nil {
		t.Errorf("Database error: %s", err.Error())
	}

	rw := httptest.NewRecorder()
	body := strings.NewReader("username=test&display_name=test&password=test")
	req := httptest.NewRequest("POST", "/signup/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postSignup(rw, req)
	if rw.Code != 303 {
		t.Errorf("Expected status 303; got %d", rw.Code)
	}

	rw = httptest.NewRecorder()
	postSignup(rw, req)
	if rw.Code != 200 {
		t.Errorf("After second signup, expected status 200 to fill out again; got %d", rw.Code)
	}
}

func TestGetLogin(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/login/", nil)
	getLogin(rw, req)
	if rw.Code != 200 {
		t.Errorf("Expected status 200; got %d", rw.Code)
	}
}

func TestPostLogin(t *testing.T) {
	rw := httptest.NewRecorder()
	body := strings.NewReader("username=test&password=test")
	req := httptest.NewRequest("POST", "/login/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postLogin(rw, req)
	if rw.Code != 303 {
		checkUser(t)
		t.Errorf("Expected status 303; got %d", rw.Code)
	}
}

func TestGetCompose(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/mail/compose/", nil)
	req.AddCookie(&http.Cookie{Name: "sessionid", Value: "1"})

	makeAuthedHandler(getCompose)(rw, req)
	if rw.Code != 200 {
		checkSession(t)
		t.Errorf("Expected status 200; got %d", rw.Code)
	}
}

func TestPostComposeSave(t *testing.T) {
	query := "delete from drafts where recipient = 'test@localhost' and user_id = 1;"
	_, err := db.Exec(query)
	if err != nil {
		t.Errorf("Database error: %s", err.Error())
	}

	rw := httptest.NewRecorder()
	body := strings.NewReader("to=test%40localhost&subject=test%20subject&content=nothing%20here")
	req := httptest.NewRequest("POST", "/mail/compose/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "sessionid", Value: "1"})

	makeAuthedHandler(postComposeSave)(rw, req)

	if rw.Code != 303 {
		checkSession(t)
		t.Errorf("Expected status 303 after first draft save; got %d", rw.Code)
	}
	rw = httptest.NewRecorder()
	makeAuthedHandler(postComposeSave)(rw, req)
	if rw.Code != 303 {
		t.Errorf("Expected status 303 after repeat draft save; got %d", rw.Code)
	}
}

func TestPostDraftSave(t *testing.T) {
	query := "delete from drafts where recipient = 'test@localhost' and user_id = 1;"
	_, err := db.Exec(query)
	if err != nil {
		t.Errorf("Database error: %s", err.Error())
	}

	rw := httptest.NewRecorder()
	body := strings.NewReader("to=test%40localhost&subject=test%20subject&content=nothing%20here")
	req := httptest.NewRequest("POST", "/mail/conv/1/save/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "sessionid", Value: "1"})

	makeAuthedHandler(postComposeSave)(rw, req)

	response := rw.Result()
	location, _ := response.Location()

	if response.StatusCode != 303 || location.Path != "/mail/conv/1/read/" {
		checkSession(t)
		checkMail(t)
		t.Errorf("Expected status 303 to '/mail/conv/1/read/' after first draft save; got status %d, '%s'", rw.Code, location.Path)
	}
	rw = httptest.NewRecorder()
	makeAuthedHandler(postComposeSave)(rw, req)
	if rw.Code != 303 {
		t.Errorf("Expected status 303 after repeat draft save; got %d", rw.Code)
	}
}

func TestPostComposeSend(t *testing.T) {
	rw := httptest.NewRecorder()
	body := strings.NewReader("to=test%40localhost&subject=test%20subject&content=nothing%20here")
	req := httptest.NewRequest("POST", "/mail/compose/send/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "sessionid", Value: "1"})

	makeAuthedHandler(postComposeSend)(rw, req)

	if rw.Code != 303 {
		checkSession(t)
		t.Errorf("Expected status 303; got %d", rw.Code)
	}
}

func TestPostDraftSend(t *testing.T) {
	rw := httptest.NewRecorder()
	body := strings.NewReader("to=test%40localhost&subject=test%20subject&content=nothing%20here")
	req := httptest.NewRequest("POST", "/mail/conv/1/send/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "sessionid", Value: "1"})

	makeAuthedHandler(postComposeSend)(rw, req)

	if rw.Code != 303 {
		checkSession(t)
		t.Errorf("Expected status 303; got %d", rw.Code)
	}
}

func TestGetInbox(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/mail/folder/inbox/", nil)
	req.AddCookie(&http.Cookie{Name: "sessionid", Value: "1"})

	makeAuthedHandler(getMailbox)(rw, req)
	if rw.Code != 200 {
		checkSession(t)
		t.Errorf("Expected status 200; got %d", rw.Code)
	}
}

func TestGetArchive(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/mail/folder/archive/", nil)
	req.AddCookie(&http.Cookie{Name: "sessionid", Value: "1"})

	makeAuthedHandler(getMailbox)(rw, req)
	if rw.Code != 200 {
		checkSession(t)
		t.Errorf("Expected status 200; got %d", rw.Code)
	}
}

func TestGetConv(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/mail/conv/1/read/", nil)
	req.AddCookie(&http.Cookie{Name: "sessionid", Value: "1"})
	req.SetPathValue("mailId", "1")

	makeAuthedHandler(getConv)(rw, req)
	if rw.Code != 200 {
		checkSession(t)
		checkMail(t)
		t.Errorf("Expected status 200; got %d", rw.Code)
	}
}

func TestLogout(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/logout/", nil)
	req.AddCookie(&http.Cookie{Name: "sessionid", Value: "1"})

	logout(rw, req)
	if rw.Code != 303 {
		t.Errorf("Expected status 303; got %d", rw.Code)
	}
}

func TestMain(m *testing.M) {
	os.Chdir("..") // tests initialize to the package directory by default
	appInit()

	os.Exit(m.Run())
}
