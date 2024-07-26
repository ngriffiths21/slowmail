package main

import "testing"

func TestLoadMail(t *testing.T) {
	_, err := loadUserMail(1)
	t.Errorf(err.Error())
}
