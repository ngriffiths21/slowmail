package main

import (
	"fmt"
	"testing"
)

func TestLoadMail(t *testing.T) {
	result, err := loadUserMail(1)
	if (err != nil) {
		t.Errorf(err.Error())
	}
	fmt.Println(result)
}
