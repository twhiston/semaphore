package router

import (
	"testing"
	"net/http"
	"strings"
	"errors"
	"encoding/json"
)

func TestGetCreateRoute(t *testing.T) {

	database := TestDb{t:t}
	route := GetCreateRoute(&CreateOptions{
		Context: "",
		NewModel:     func() interface{}{return map[string]string{"test": "test"}},
		ProcessInput: func(context interface{}, model interface{}) error {
			t.Log("Input processed")
			return nil
		},
	},
	&database)

	w := TestResponseWriter{t: t}

	//Empty input should cause a fail
	r, err := http.NewRequest("GET","/projects", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	route(&w,r)

	if w.status != 400 {
		t.Fail()
	}

	//In this test case any valid data input is regarded as written correctly
	r, err = http.NewRequest("GET","/projects", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	route(&w,r)

	if w.status != 201 {
		t.Fail()
	}

}

func TestFailingValidationPath(t *testing.T) {
	database := TestDb{t:t}
	route := GetCreateRoute(&CreateOptions{
		Context: "",
		NewModel:     func() interface{}{return map[string]string{"test": "test"}},
		ProcessInput: func(context interface{}, model interface{}) error {
			return errors.New("boom")
		},

	}, &database)

	//now call the route
	w := TestResponseWriter{t: t}

	r, err := http.NewRequest("GET","/projects", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	route(&w,r)
	//This should fail on the validation state
	if w.status != 400 {
		t.Fail()
	}
	body := make(map[string]string)
	err = json.Unmarshal(w.input, &body)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := body["error"]; if !ok {
		t.Fatal("should exit due to invalid json error but no error key exists")
	}
}