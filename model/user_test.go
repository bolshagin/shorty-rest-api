package model_test

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var (
	client   = http.Client{Timeout: 5 * time.Second}
	email    = "example@example.ru"
	password = "123456"
)

func TestCreateUser(t *testing.T) {
	data := []byte(fmt.Sprintf(`{"email":"%v","password":"%v"}`, email, password))

	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/users", bytes.NewBuffer(data))
	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	defer req.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestCreateUserAgain(t *testing.T) {
	data := []byte(fmt.Sprintf(`{"email":"%v","password":"%v"}`, email, password))

	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/users", bytes.NewBuffer(data))
	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	defer req.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
