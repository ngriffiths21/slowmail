package main

import (
	"testing"
	"net/http/httptest"
	"os"
	"net/http"
	"strings"
)

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
		t.Errorf("Expected status 303; got %d", rw.Code)
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
		t.Errorf("Expected status 303 after first draft save; got %d", rw.Code)
	}
	rw = httptest.NewRecorder()
	makeAuthedHandler(postComposeSave)(rw, req)
	if rw.Code != 303 {
		t.Errorf("Expected status 303 after repeat draft save; got %d", rw.Code)
	}
}

func TestGetConv(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/mail/conv/1/read/", nil)
	makeAuthedHandler(getConv)(rw, req)
	if rw.Code != 200 {
		t.Errorf("Expected status 200; got %d", rw.Code)
	}
}

func TestMain(m *testing.M) {
	os.Chdir("..") // tests initialize to the package directory by default
	appInit()
	os.Exit(m.Run())
}
