package model_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestCreateUser(t *testing.T) {
	data := []byte(`{"email":"example@example.ru","password":"123456"}`)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	req, _ := http.NewRequest("POST", "http://localhost:8080/users", bytes.NewBuffer(data))
	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}

	var m map[string]interface{}
	json.Unmarshal(body, &m)

	assert.Equal(t, "example@example.ru", m["email"])
}
